wt() {
  local result
  result=$(command wt --output-path "$@")
  local exit_code=$?

  if [ $exit_code -eq 0 ] && [ -n "$result" ]; then
    cd "$result" || return 1
  fi

  return $exit_code
}
