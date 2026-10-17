#!/bin/bash
# Cal.com reuses the GET and PATCH calendar event operation IDs on both
# /events/{eventUid} and /event/{eventUid}. oapi-codegen emits duplicate Go
# declarations. Name the singular-path aliases separately, preserving paths,
# methods, parameters and schemas for all operations.
# ReassignBookingOutput.data refers to its own envelope through both oneOf
# and allOf, producing an invalid recursive Go value. Its description names
# two response types absent from the spec. Use an open object for data until
# the provider supplies those schemas, preserving all returned JSON fields.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

main() {
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.openapi.patched.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp before the local leaves scope
	trap "rm -f -- '${tmp}'" EXIT

	jq '
      .paths["/v2/calendars/{calendar}/event/{eventUid}"] |=
        with_entries(
          if .key == "get" or .key == "patch" then
            .value.operationId += "_singularPath"
          else . end
        )
      | .components.schemas.ReassignBookingOutput_2024_08_13
          .properties.data |=
          (del(.oneOf, .allOf) +
            {type: "object", additionalProperties: true})
    ' "${SPEC_DIR}/openapi.json" >"$tmp"

	# Fail on new duplicate IDs rather than emitting an uncompilable client.
	jq -e '
      [.paths[][] | objects | .operationId? // empty] as $ids
      | ($ids | length) == ($ids | unique | length)
    ' "$tmp" >/dev/null

	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/openapi.patched.json"
	log INFO "patched duplicate operation IDs and recursive reassignment data"
}

main "$@"
