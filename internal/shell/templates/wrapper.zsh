wt() {
  case "$1" in
    goto|home|new|eject)
      local result
      result=$("__WT_BIN__" --output-path "$@")
      local exit_code=$?
      if [ $exit_code -eq 0 ] && [ -n "$result" ]; then
        cd "$result" || return 1
      fi
      return $exit_code
      ;;
    *)
      "__WT_BIN__" "$@"
      ;;
  esac
}

# Completions
eval "$("__WT_BIN__" completion zsh 2>/dev/null)"
