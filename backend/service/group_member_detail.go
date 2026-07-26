package service

// group_member_detail.go — 「群友画像」：单个群成员在某个群里的发言分析
//
// Issue #104：群聊 tab 里只有好友才能点开分析，且分析以私聊为主。
// 这里提供按群维度的成员分析 —— 无论对方是不是好友，只要在群里
// 发过言（real_sender_id → Name2Id → wxid），就能算出 TA 的群内画像。
//
// 一次全表扫描：对目标成员做精细聚合（时段/词云/类型/最近发言），
// 同时顺手统计所有人的发言数，得出目标的群内排名和占比。

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"welink/backend/pkg/db"
)

const (
	gmdTopWords      = 30   // 高频词 top N
	gmdRecentMsgs    = 6    // 最近发言样本条数
	gmdMaxSegSamples = 6000 // 分词文本样本上限（防超大群成员拖垮）
	gmdMaxSampleRune = 120  // 单条样本最长字符
	gmdTrendMonths   = 12   // 月度趋势最近 N 个月
)

// GMDMsgSample 一条最近发言样本
type GMDMsgSample struct {
	Time    string `json:"time"` // "2024-03-15 14:23"
	Content string `json:"content"`
}

// GMDMonthCount 月度发言数
type GMDMonthCount struct {
	Month string `json:"month"` // "2024-03"
	Count int    `json:"count"`
}

// GroupMemberDetail 群成员画像
type GroupMemberDetail struct {
	Speaker   string `json:"speaker"`    // 显示名
	Username  string `json:"username"`   // wxid
	AvatarURL string `json:"avatar_url"` // 头像（联系人表里有才有）
	IsContact bool   `json:"is_contact"` // 是否在联系人表（= 好友/可跳私聊分析）

	Count            int64  `json:"count"`      // 全类型消息数
	TextCount        int64  `json:"text_count"` // 文本消息数
	FirstMessageTime string `json:"first_message_time"`
	LastMessageTime  string `json:"last_message_time"`
	FirstMessageTs   int64  `json:"first_message_ts"`
	LastMessageTs    int64  `json:"last_message_ts"`
	ActiveDays       int    `json:"active_days"` // 有发言的天数

	Rank         int     `json:"rank"`           // 群内发言排名（1 起）
	TotalSpoken  int     `json:"total_spoken"`   // 群内发过言的人数
	SharePct     float64 `json:"share_pct"`      // 占全群消息百分比
	GroupTotal   int64   `json:"group_total"`    // 全群消息总数
	LateNightCnt int64   `json:"late_night_cnt"` // 深夜发言数（0~5 点）

	HourlyDist   [24]int         `json:"hourly_dist"`
	WeeklyDist   [7]int          `json:"weekly_dist"`
	MonthlyTrend []GMDMonthCount `json:"monthly_trend"`
	TypeDist     map[string]int  `json:"type_dist"`
	TopWords     []WordCount     `json:"top_words"`
	RecentMsgs   []GMDMsgSample  `json:"recent_msgs"`
}

// GetGroupMemberDetail 计算某成员（wxid）在某群的发言画像。
func (s *ContactService) GetGroupMemberDetail(groupUsername, memberWxid string) (*GroupMemberDetail, error) {
	if !strings.HasSuffix(groupUsername, "@chatroom") {
		return nil, fmt.Errorf("username 必须是群聊（以 @chatroom 结尾）")
	}
	if memberWxid == "" {
		return nil, fmt.Errorf("member 不能为空")
	}

	tableName := db.GetTableName(groupUsername)
	tw := s.timeWhere()

	d := &GroupMemberDetail{
		Username: memberWxid,
		TypeDist: make(map[string]int),
		TopWords: []WordCount{},
	}
	// 成员名 / 头像：单行点查，避免为一个人全表扫 contact 两遍
	d.Speaker = memberWxid
	if s.dbMgr != nil && s.dbMgr.ContactDB != nil {
		var remark, nick, avatar string
		if err := s.dbMgr.ContactDB.QueryRow(
			`SELECT COALESCE(remark,''), COALESCE(nick_name,''), COALESCE(small_head_url,'') FROM contact WHERE username = ?`,
			memberWxid).Scan(&remark, &nick, &avatar); err == nil {
			d.IsContact = true
			if remark != "" {
				d.Speaker = remark
			} else if nick != "" {
				d.Speaker = nick
			}
			d.AvatarURL = avatar
		}
	}

	perMember := make(map[string]int64) // wxid → 发言数（算排名/占比）
	monthly := make(map[string]int)
	daySet := make(map[string]struct{})
	var textSamples []string
	type recentEntry struct {
		ts      int64
		content string
	}
	var recent []recentEntry

	for _, mdb := range s.msgRepo.DBsForUsername(groupUsername) {
		// 本 DB 的 rowid ↔ wxid，同时记下目标成员在本 DB 的 rowid
		idToWxid := make(map[int64]string)
		var memberRowIDs []int64
		if nrows, nerr := mdb.Query("SELECT rowid, user_name FROM Name2Id"); nerr == nil {
			for nrows.Next() {
				var rid int64
				var uname string
				nrows.Scan(&rid, &uname)
				idToWxid[rid] = uname
				if uname == memberWxid {
					memberRowIDs = append(memberRowIDs, rid)
				}
			}
			nrows.Close()
		}

		// Pass A：全群按发言人聚合条数（不解码内容，很便宜）→ 排名 / 占比 / 总数
		if arows, aerr := mdb.Query(fmt.Sprintf(
			"SELECT real_sender_id, COUNT(*) FROM [%s]%s GROUP BY real_sender_id", tableName, tw)); aerr == nil {
			for arows.Next() {
				var rid, cnt int64
				if arows.Scan(&rid, &cnt) != nil {
					continue
				}
				d.GroupTotal += cnt
				if wxid := idToWxid[rid]; wxid != "" {
					perMember[wxid] += cnt
				}
			}
			arows.Close()
		}

		// Pass B：只查目标成员自己的消息，解码内容只发生在 TA 的行上
		if len(memberRowIDs) == 0 {
			continue
		}
		ids := make([]string, len(memberRowIDs))
		for i, r := range memberRowIDs {
			ids[i] = strconv.FormatInt(r, 10)
		}
		where := " WHERE real_sender_id IN (" + strings.Join(ids, ",") + ")"
		if tw != "" {
			where += " AND (" + strings.TrimPrefix(tw, " WHERE ") + ")"
		}
		rows, err := mdb.Query(fmt.Sprintf(
			"SELECT create_time, local_type, message_content, COALESCE(WCDB_CT_message_content,0) FROM [%s]%s",
			tableName, where))
		if err != nil {
			continue
		}
		for rows.Next() {
			var ts int64
			var lt int
			var rawContent []byte
			var ct int64
			rows.Scan(&ts, &lt, &rawContent, &ct)

			content := decodeGroupContent(rawContent, ct)
			dt := time.Unix(ts, 0).In(s.tz)
			d.Count++
			d.HourlyDist[dt.Hour()]++
			d.WeeklyDist[int(dt.Weekday())]++
			if dt.Hour() < 6 {
				d.LateNightCnt++
			}
			monthly[dt.Format("2006-01")]++
			daySet[dt.Format("2006-01-02")] = struct{}{}
			if typeName := classifyMsgType(lt, content); typeName != "系统" {
				d.TypeDist[typeName]++
			}
			if ts > d.LastMessageTs {
				d.LastMessageTs = ts
			}
			if d.FirstMessageTs == 0 || ts < d.FirstMessageTs {
				d.FirstMessageTs = ts
			}

			if (lt&0xFFFF) == 1 && content != "" {
				txt := content
				// 群消息可能带 "wxid:\n" 前缀，剥掉
				if idx := strings.Index(txt, ":\n"); idx > 0 && idx < 80 {
					txt = txt[idx+2:]
				}
				if txt == "" || s.isSys(txt) {
					continue
				}
				d.TextCount++
				if len(textSamples) < gmdMaxSegSamples {
					textSamples = append(textSamples, wechatEmojiRe.ReplaceAllString(txt, ""))
				}
				// 最近发言样本：维护 ts 最大的 N 条
				sample := txt
				if rs := []rune(sample); len(rs) > gmdMaxSampleRune {
					sample = string(rs[:gmdMaxSampleRune]) + "…"
				}
				if len(recent) < gmdRecentMsgs {
					recent = append(recent, recentEntry{ts, sample})
				} else {
					minIdx := 0
					for i := 1; i < len(recent); i++ {
						if recent[i].ts < recent[minIdx].ts {
							minIdx = i
						}
					}
					if ts > recent[minIdx].ts {
						recent[minIdx] = recentEntry{ts, sample}
					}
				}
			}
		}
		rows.Close()
	}

	if d.Count == 0 {
		// 没发过言（潜水/不在 Name2Id）—— 返回空画像而不是错误，前端展示"暂无发言"
		return d, nil
	}

	d.ActiveDays = len(daySet)
	if d.FirstMessageTs > 0 {
		d.FirstMessageTime = time.Unix(d.FirstMessageTs, 0).In(s.tz).Format("2006-01-02 15:04")
	}
	if d.LastMessageTs > 0 {
		d.LastMessageTime = time.Unix(d.LastMessageTs, 0).In(s.tz).Format("2006-01-02 15:04")
	}
	if d.GroupTotal > 0 {
		d.SharePct = float64(d.Count) / float64(d.GroupTotal) * 100
	}

	// 排名：发言数比目标多的人数 + 1
	d.TotalSpoken = len(perMember)
	rank := 1
	for _, c := range perMember {
		if c > d.Count {
			rank++
		}
	}
	d.Rank = rank

	// 月度趋势：取最近 N 个月（按自然月连续铺开，没数据补 0）
	if len(monthly) > 0 {
		end := time.Unix(d.LastMessageTs, 0).In(s.tz)
		for i := gmdTrendMonths - 1; i >= 0; i-- {
			m := end.AddDate(0, -i, 0).Format("2006-01")
			d.MonthlyTrend = append(d.MonthlyTrend, GMDMonthCount{Month: m, Count: monthly[m]})
		}
	}

	// 高频词（gse 非线程安全，分批加锁，同 computeGroupDetail）
	wordCounts := make(map[string]int)
	const segBatch = 500
	for i := 0; i < len(textSamples); i += segBatch {
		end := i + segBatch
		if end > len(textSamples) {
			end = len(textSamples)
		}
		s.segmenterMu.Lock()
		for _, text := range textSamples[i:end] {
			for _, seg := range s.segmenter.Cut(text, true) {
				seg = strings.TrimSpace(seg)
				if !utf8.ValidString(seg) {
					continue
				}
				runes := []rune(seg)
				if len(runes) < 2 || len(runes) > 8 {
					continue
				}
				if isNumeric(seg) || STOP_WORDS[seg] || containsEmoji(seg) || !hasWordChar(seg) {
					continue
				}
				wordCounts[seg]++
			}
		}
		s.segmenterMu.Unlock()
	}
	for w, c := range wordCounts {
		d.TopWords = append(d.TopWords, WordCount{w, c})
	}
	sort.Slice(d.TopWords, func(i, j int) bool { return d.TopWords[i].Count > d.TopWords[j].Count })
	if len(d.TopWords) > gmdTopWords {
		d.TopWords = d.TopWords[:gmdTopWords]
	}

	// 最近发言按时间倒序
	sort.Slice(recent, func(i, j int) bool { return recent[i].ts > recent[j].ts })
	d.RecentMsgs = make([]GMDMsgSample, 0, len(recent))
	for _, r := range recent {
		d.RecentMsgs = append(d.RecentMsgs, GMDMsgSample{
			Time:    time.Unix(r.ts, 0).In(s.tz).Format("2006-01-02 15:04"),
			Content: r.content,
		})
	}

	return d, nil
}
