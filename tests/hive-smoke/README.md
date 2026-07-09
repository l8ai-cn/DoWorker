# Hive smoke scripts

API / browser smoke suites for the hive (web-user) stack.

Run via:

```bash
bash deploy/dev/hive_smoke.sh
# or
bazel test //deploy/dev:hive_smoke --test_tag_filters=hive,manual
```

Scripts write screenshots and JSON reports under repo-root `output/` (gitignored scratch).
