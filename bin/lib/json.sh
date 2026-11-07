#!/usr/bin/env bash
# json helpers shared by the bin/* entrypoints that grew a --json mode.
#
# Why this exists: nekono-pacman-agent (the harness) must not parse human log
# text. Every entrypoint that the agent calls therefore prints one JSON object
# on stdout when given --json, and keeps its existing text output otherwise.
# In --json mode the human text is redirected to stderr, so child process logs
# still stream while stdout carries exactly one machine-readable document.
#
# Keep this file POSIX-bash only: it is sourced by scripts that run on a bare
# Arch cloud image before anything else is installed.

# json_escape STRING -> escaped body (no surrounding quotes)
json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\n'/\\n}"
    s="${s//$'\r'/\\r}"
    s="${s//$'\t'/\\t}"
    printf '%s' "$s"
}

# json_str STRING -> quoted JSON string
json_str() {
    printf '"%s"' "$(json_escape "$1")"
}

# json_true / json_false from a shell boolean-ish value (true/1/yes -> true)
json_bool() {
    case "${1:-}" in
        true | 1 | yes | ok | OK) printf 'true' ;;
        *) printf 'false' ;;
    esac
}

# json_str_list a b c -> ["a","b","c"]
json_str_list() {
    local out="" x
    for x in "$@"; do
        out+="${out:+,}$(json_str "$x")"
    done
    printf '[%s]' "$out"
}

# json_lines "multi line string" -> ["l1","l2"] (empty lines dropped)
json_lines() {
    local out="" line
    while IFS= read -r line; do
        [ -n "$line" ] || continue
        out+="${out:+,}$(json_str "$line")"
    done <<< "${1:-}"
    printf '[%s]' "$out"
}

# json_pairs "pkg" "ver" [ "pkg" "ver" ... ] -> [{"pkg":"pkg","version":"ver"}]
json_pkg_versions() {
    local out=""
    while [ "$#" -gt 0 ]; do
        out+="${out:+,}{\"pkg\":$(json_str "$1"),\"version\":$(json_str "${2:-}")}"
        shift 2 || break
    done
    printf '[%s]' "$out"
}
