#!/usr/bin/env bash

set -eux

: "${VENOM_VAR_binpath:? Please set VENOM_VAR_binpath to updatecli binary dirname}"
: "${VENOM_VAR_rootpath:=../..}"

## In order for the Updatecli manifest to work, we must be at the root of the git repository
pushd "$VENOM_VAR_rootpath" || exit

# GitLab pipelines call the real, unauthenticated gitlab.com API and are the only ones in
# this suite exposed to its anonymous rate limit. They're run on their own below so a
# rate-limited/unreachable gitlab.com can be treated as a skip instead of failing every
# other, unrelated pipeline in this run.
gitlab_dir="e2e/updatecli.d/success.d/gitlab"
gitlab_tmp="$(mktemp -d)/gitlab"
mv "$gitlab_dir" "$gitlab_tmp"
trap 'mv "$gitlab_tmp" "$gitlab_dir" 2>/dev/null || true' EXIT

echo "+++++++++++++++++++++++"
echo "+ GITLAB PIPELINE(S)   +"
echo "+++++++++++++++++++++++"

set +e
gitlab_output="$("$VENOM_VAR_binpath/updatecli" pipeline diff --config "$gitlab_tmp" 2>&1)"
gitlab_code=$?
set -e

if [ "$gitlab_code" -ne 0 ] && printf '%s' "$gitlab_output" | grep -qE 'rate: Wait\([^)]*\) would exceed context deadline|dial tcp[^ ]* i/o timeout|dial tcp: lookup|no such host|connection refused|429 Too Many Requests'; then
	echo "::warning::GitLab e2e pipelines skipped: gitlab.com looks rate-limited or unreachable from this runner"
	gitlab_code=0
else
	echo "$gitlab_output"
fi

echo "+++++++++++++++++++++++"
echo "+ REMAINING PIPELINES  +"
echo "+++++++++++++++++++++++"

set +e
"$VENOM_VAR_binpath/updatecli" pipeline diff --config e2e/updatecli.d/success.d
main_code=$?
set -e

if [ "$gitlab_code" -ne 0 ]; then
	exit "$gitlab_code"
fi
exit "$main_code"
