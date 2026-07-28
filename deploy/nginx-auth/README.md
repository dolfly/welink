# Nginx Basic Auth

Production traffic reports at `/_stats/` require a local `htpasswd` file in this directory.

Do not commit the generated `htpasswd` file.

Example on the production server:

```bash
apt-get install -y apache2-utils
htpasswd -c /opt/welink/deploy/nginx-auth/htpasswd <username>
docker compose -f /opt/welink/server-compose.yml restart website
```
