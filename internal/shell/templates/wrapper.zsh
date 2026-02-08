ptt() {
  case "$1" in
    go|goto|home|mk|new|eject)
      local result
      result=$("__PTT_BIN__" --output-path "$@")
      local exit_code=$?
      if [ $exit_code -eq 0 ] && [ -n "$result" ]; then
        cd "$result" || return 1
      fi
      return $exit_code
      ;;
    *)
      "__PTT_BIN__" "$@"
      ;;
  esac
}

# Completions
eval "$("__PTT_BIN__" completion zsh 2>/dev/null)"
