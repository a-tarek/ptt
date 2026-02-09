function ptt
  switch $argv[1]
    case cd mk new eject
      set -l result ("__PTT_BIN__" --output-path $argv)
      set -l exit_code $status
      if test $exit_code -eq 0 -a -n "$result"
        cd "$result"; or return 1
      end
      return $exit_code
    case '*'
      "__PTT_BIN__" $argv
  end
end

# Completions
"__PTT_BIN__" completion fish 2>/dev/null | source
