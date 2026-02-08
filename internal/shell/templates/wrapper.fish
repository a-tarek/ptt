function wt
  switch $argv[1]
    case goto home new eject
      set -l result ("__WT_BIN__" --output-path $argv)
      set -l exit_code $status
      if test $exit_code -eq 0 -a -n "$result"
        cd "$result"; or return 1
      end
      return $exit_code
    case '*'
      "__WT_BIN__" $argv
  end
end

# Completions
"__WT_BIN__" completion fish 2>/dev/null | source
