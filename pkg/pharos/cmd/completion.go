package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd implements a CLI command that allows a user to register into
// their shell of choice (bash or zsh) autocomplete functionality for the Pharos
// command.
//
// The following cobra.Command is heavily derived from the Kubernetes kubectl
// autocomplete functionality:
// https://github.com/kubernetes/kubernetes/blob/c30f0248649c15a46ab99eb722de0448988198f8/pkg/kubectl/cmd/completion/completion.go
var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Generates completion scripts for the specified shell (bash or zsh), defaulting to bash",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case len(args) == 0:
			return rootCmd.GenBashCompletion(os.Stdout)
		case args[0] == "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case args[0] == "zsh":
			return runCompletionZsh(os.Stdout)
		default:
			return fmt.Errorf("%q is not a supported shell", args[0])
		}
	},
}

func runCompletionZsh(out io.Writer) error {
	var b bytes.Buffer
	buf := bufio.NewWriter(&b)

	buf.WriteString(zshInitialization) //nolint
	rootCmd.GenBashCompletion(buf)     //nolint
	buf.WriteString(zshTail)           //nolint
	buf.Flush()                        //nolint

	out.Write(b.Bytes()) //nolint

	return nil
}

var (
	zshInitialization = `
#compdef pharos
__pharos_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__pharos_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__pharos_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__pharos_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__pharos_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__pharos_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__pharos_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__pharos_filedir() {
	local RET OLD_IFS w qw
	__pharos_debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __pharos_debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__pharos_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__pharos_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
	printf %q "$1"
    fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__pharos_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__pharos_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__pharos_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__pharos_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__pharos_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__pharos_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__pharos_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	zshTail = `
BASH_COMPLETION_EOF
}
__pharos_bash_source <(__pharos_convert_bash_to_zsh)
_complete pharos 2>/dev/null
`
)
