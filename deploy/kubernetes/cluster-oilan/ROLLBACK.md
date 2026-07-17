# AgentsMesh Oilan Rollback

1. Revert the release commit on `main`, push it, and wait for the required
   GitHub checks. This restores the previous GitOps manifests and image digests.
2. Stop application writes before restoring a database:

```bash
doops -session <rollback-session> exec --target gw-oilan-node --cmd \
  'kubectl -n agentsmesh scale deploy/backend deploy/marketplace --replicas=0 &&
   kubectl -n agentsmesh wait --for=delete pod -l app=backend --timeout=180s &&
   kubectl -n agentsmesh wait --for=delete pod -l app=marketplace --timeout=180s'
```

3. Select the exact pre-release backup printed by the failed deployment. Never
   use a moving `latest` symlink. Verify its checksum, terminate database
   sessions, recreate the database, and restore:

```bash
doops -session <rollback-session> exec --target gw-oilan-node --cmd \
  'set -eu; cd /root/backups/agentsmesh;
   backup="<pre-migrate-release-sha-UTC>.dump";
   test -f "$backup" && test -f "$backup.sha256";
   sha256sum -c "$backup.sha256";
   kubectl -n agentsmesh exec deploy/postgres -- sh -ceu '"'"'
     export PGPASSWORD="$POSTGRES_PASSWORD";
     dropdb --force --if-exists -U "$POSTGRES_USER" "$POSTGRES_DB";
     createdb -U "$POSTGRES_USER" "$POSTGRES_DB"'"'"';
   cat "$backup" | kubectl -n agentsmesh exec -i deploy/postgres -- sh -ceu '"'"'
     export PGPASSWORD="$POSTGRES_PASSWORD";
     pg_restore --clean --if-exists --no-owner --no-privileges \
       -U "$POSTGRES_USER" -d "$POSTGRES_DB"'"'"
```

4. Run `deploy.sh` from the pushed revert commit. Complete health and browser
   smoke checks before reopening traffic.
