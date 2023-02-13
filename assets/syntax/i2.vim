" Comment
syntax region	Comment start="/\*" end="\*/"

" Constant
syntax region	String	start=+"+ skip=+\\"+ end=+"+

" Identifier


" Statement
syntax keyword	Keyword	        mod import export
syntax keyword	Repeat	        for some with
syntax keyword	Type	        func type
syntax keyword	Label	        this
syntax keyword	Exception	true false

syntax match	logicOp	        display	"===\|==>\|<==\|==\|!=\|!"

" Type
syntax keyword	StorageClass	auto

" Special

