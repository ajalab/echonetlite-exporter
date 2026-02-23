#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  mra_query.sh list-props <device> [--mra-dir <dir>] [--format json|table]
  mra_query.sh prop-def <device> <property_short_name> [--mra-dir <dir>] [--resolve-data] [--format json|pretty]

Examples:
  mra_query.sh list-props storageBattery
  mra_query.sh list-props 0x027D --format table
  mra_query.sh prop-def 0x027D acDischargeableCapacity
  mra_query.sh prop-def 0x027D acDischargeableCapacity --resolve-data
EOF
}

err() {
  echo "Error: $*" >&2
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    err "Required command not found: $cmd"
    exit 1
  fi
}

normalize_eoj() {
  local value="$1"
  printf '0x%s' "${value#0x}" | tr '[:lower:]' '[:upper:]'
}

validate_mra_dir() {
  local mra_dir="$1"
  [[ -d "$mra_dir/devices" ]] || { err "Missing directory: $mra_dir/devices"; exit 1; }
  [[ -d "$mra_dir/superClass" ]] || { err "Missing directory: $mra_dir/superClass"; exit 1; }
  [[ -d "$mra_dir/nodeProfile" ]] || { err "Missing directory: $mra_dir/nodeProfile"; exit 1; }
  [[ -f "$mra_dir/definitions/definitions.json" ]] || { err "Missing file: $mra_dir/definitions/definitions.json"; exit 1; }
}

resolve_device_file() {
  local mra_dir="$1"
  local device_input="$2"
  local candidates=()
  local mode="shortName"
  local normalized=""
  local file

  shopt -s nullglob
  for file in "$mra_dir"/devices/*.json "$mra_dir"/superClass/*.json "$mra_dir"/nodeProfile/*.json; do
    if [[ "$device_input" =~ ^0[xX][0-9A-Fa-f]{4}$ ]]; then
      mode="eoj"
      normalized="$(normalize_eoj "$device_input")"
      if jq -e --arg eoj "$normalized" '.eoj | ascii_upcase == ($eoj | ascii_upcase)' "$file" >/dev/null 2>&1; then
        candidates+=("$file")
      fi
    else
      if jq -e --arg short "$device_input" '.shortName == $short' "$file" >/dev/null 2>&1; then
        candidates+=("$file")
      fi
    fi
  done
  shopt -u nullglob

  if [[ ${#candidates[@]} -eq 0 ]]; then
    if [[ "$mode" == "eoj" ]]; then
      err "No device matched EOJ '$normalized' in $mra_dir"
    else
      err "No device matched shortName '$device_input' in $mra_dir"
    fi
    exit 1
  fi

  if [[ ${#candidates[@]} -gt 1 ]]; then
    err "Multiple device matches for '$device_input':"
    printf '  %s\n' "${candidates[@]}" >&2
    exit 1
  fi

  printf '%s\n' "${candidates[0]}"
}

list_props_json() {
  local device_file="$1"
  jq '{
    device: {
      eoj,
      shortName,
      className
    },
    properties: [
      .elProperties[] | {
        epc,
        shortName,
        propertyName,
        accessRule,
        validRelease
      }
    ]
  }' "$device_file"
}

list_props_table() {
  local device_file="$1"
  local header
  header="$(jq -r '[.eoj, .shortName, (.className.en // "")] | @tsv' "$device_file")"
  IFS=$'\t' read -r eoj short_name class_en <<<"$header"
  printf 'Device: %s (%s)%s\n' "$short_name" "$eoj" "${class_en:+ - $class_en}"
  printf '%-6s %-38s %-12s %-12s %-12s %s\n' "EPC" "shortName" "GET" "SET" "INF" "propertyName.en"
  jq -r '
    .elProperties[]
    | [
        (.epc // ""),
        (.shortName // ""),
        (.accessRule.get // ""),
        (.accessRule.set // ""),
        (.accessRule.inf // ""),
        (.propertyName.en // "")
      ]
    | @tsv
  ' "$device_file" | while IFS=$'\t' read -r epc prop get set inf en; do
    printf '%-6s %-38s %-12s %-12s %-12s %s\n' "$epc" "$prop" "$get" "$set" "$inf" "$en"
  done
}

prop_def_json() {
  local device_file="$1"
  local property_short="$2"
  jq --arg prop "$property_short" '
    .elProperties[] | select(.shortName == $prop)
  ' "$device_file"
}

property_exists() {
  local device_file="$1"
  local property_short="$2"
  jq -e --arg prop "$property_short" 'any(.elProperties[]; .shortName == $prop)' "$device_file" >/dev/null
}

prop_def_json_with_resolve() {
  local device_file="$1"
  local defs_file="$2"
  local property_short="$3"
  jq -e -n \
    --slurpfile dev "$device_file" \
    --slurpfile defs "$defs_file" \
    --arg prop "$property_short" '
      def resolve_refs($defs):
        if type == "object" then
          if ((keys | length) == 1 and has("$ref") and (."$ref" | type == "string") and (."$ref" | startswith("#/definitions/"))) then
            (."$ref" | ltrimstr("#/definitions/")) as $key
            | ($defs[0].definitions[$key] // .)
          else
            with_entries(.value |= resolve_refs($defs))
          end
        elif type == "array" then
          map(resolve_refs($defs))
        else
          .
        end;

      ($dev[0].elProperties[] | select(.shortName == $prop)) as $p
      | {
          device: {
            eoj: $dev[0].eoj,
            shortName: $dev[0].shortName,
            className: $dev[0].className
          },
          property: $p,
          dataRef: ($p.data."$ref" // null),
          resolvedData: (
            if ($p.data | type) != "object" then null
            elif ($p.data | has("$ref") | not) then null
            else ($p.data | resolve_refs($defs))
            end
          )
        }
    '
}

prop_not_found_error() {
  local device_file="$1"
  local property_short="$2"
  local device_summary
  device_summary="$(jq -r '[.shortName, .eoj] | @tsv' "$device_file")"
  IFS=$'\t' read -r short_name eoj <<<"$device_summary"
  err "Property shortName '$property_short' not found for device '$short_name' ($eoj)"
  err "Hint: run 'mra_query.sh list-props $short_name' to inspect valid property short names."
}

main() {
  require_cmd jq

  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  local subcmd="$1"
  shift

  local mra_dir="MRA_v1.3.1"
  local format=""
  local resolve_data="false"
  local device=""
  local property_short=""
  local device_file=""
  local defs_file=""

  case "$subcmd" in
    list-props)
      if [[ $# -lt 1 ]]; then
        usage
        exit 1
      fi
      device="$1"
      shift
      format="json"
      while [[ $# -gt 0 ]]; do
        case "$1" in
          --mra-dir)
            [[ $# -ge 2 ]] || { err "--mra-dir requires a value"; exit 1; }
            mra_dir="$2"
            shift 2
            ;;
          --format)
            [[ $# -ge 2 ]] || { err "--format requires a value"; exit 1; }
            format="$2"
            shift 2
            ;;
          -h|--help)
            usage
            exit 0
            ;;
          *)
            err "Unknown argument for list-props: $1"
            usage
            exit 1
            ;;
        esac
      done
      validate_mra_dir "$mra_dir"
      device_file="$(resolve_device_file "$mra_dir" "$device")"
      case "$format" in
        json) list_props_json "$device_file" ;;
        table) list_props_table "$device_file" ;;
        *)
          err "Unsupported format for list-props: $format (use json or table)"
          exit 1
          ;;
      esac
      ;;

    prop-def)
      if [[ $# -lt 2 ]]; then
        usage
        exit 1
      fi
      device="$1"
      property_short="$2"
      shift 2
      format="json"
      while [[ $# -gt 0 ]]; do
        case "$1" in
          --mra-dir)
            [[ $# -ge 2 ]] || { err "--mra-dir requires a value"; exit 1; }
            mra_dir="$2"
            shift 2
            ;;
          --resolve-data)
            resolve_data="true"
            shift
            ;;
          --format)
            [[ $# -ge 2 ]] || { err "--format requires a value"; exit 1; }
            format="$2"
            shift 2
            ;;
          -h|--help)
            usage
            exit 0
            ;;
          *)
            err "Unknown argument for prop-def: $1"
            usage
            exit 1
            ;;
        esac
      done
      validate_mra_dir "$mra_dir"
      device_file="$(resolve_device_file "$mra_dir" "$device")"
      defs_file="$mra_dir/definitions/definitions.json"

      case "$format" in
        json|pretty) ;;
        *)
          err "Unsupported format for prop-def: $format (use json or pretty)"
          exit 1
          ;;
      esac

      if ! property_exists "$device_file" "$property_short"; then
        prop_not_found_error "$device_file" "$property_short"
        exit 1
      fi

      if [[ "$resolve_data" == "true" ]]; then
        prop_def_json_with_resolve "$device_file" "$defs_file" "$property_short"
      else
        prop_def_json "$device_file" "$property_short"
      fi
      ;;

    -h|--help|help)
      usage
      ;;

    *)
      err "Unknown subcommand: $subcmd"
      usage
      exit 1
      ;;
  esac
}

main "$@"
