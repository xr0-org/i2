" Comment
syntax region	Comment start="/\*" end="\*/"

" Constant
syntax region	String	start=+"+ skip=+\\"+ end=+"+

" Identifier


" Statement
syntax keyword	Keyword	        mod import export bool
syntax keyword	Repeat	        for some with
syntax keyword	Type	        func tmpl
syntax keyword	Label	        this
syntax keyword	Exception	true false

syntax match	logicOp	        display	"===\|==>\|<==\|==\|!=\|!"

" Type
syntax keyword	StorageClass	auto

" Special

