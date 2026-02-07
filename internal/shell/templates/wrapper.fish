function wt
  set -l result (command wt --output-path $argv)
  set -l exit_code $status

  if test $exit_code -eq 0 -a -n "$result"
    cd "$result"; or return 1
  end

  return $exit_code
end
