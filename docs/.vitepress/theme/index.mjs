import DefaultTheme from 'vitepress/theme'
import { onMounted, watch, nextTick } from 'vue'
import { useRoute } from 'vitepress'
import mediumZoom from 'medium-zoom'
import './custom.css'

export default {
  extends: DefaultTheme,
  setup() {
    const route = useRoute()

    const initZoom = () => {
      // 对所有文档区域的 img 启用 zoom，排除 logo 等 UI 图片
      mediumZoom('.vp-doc img', {
        background: 'rgba(0,0,0,0.8)',
        margin: 24,
      })
    }

    onMounted(() => initZoom())
    watch(() => route.path, () => nextTick(() => initZoom()))
  },
}
