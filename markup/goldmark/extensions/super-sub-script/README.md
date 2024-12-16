[![Documentation](https://pkg.go.dev/badge/github.com/bowman2001/goldmark-supersubscript.svg)](https://pkg.go.dev/github.com/bowman2001/goldmark-supersubscript)
[![Test Status](https://github.com/bowman2001/goldmark-supersubscript/workflows/test/badge.svg)](https://github.com/bowman2001/goldmark-supersubscript/actions?query=workflow:test)
[![Coverage Status](https://coveralls.io/repos/github/bowman2001/goldmark-supersubscript/badge.svg)](https://coveralls.io/github/bowman2001/goldmark-supersubscript)
[![Report Card](https://goreportcard.com/badge/github.com/bowman2001/goldmark-supersubscript)](https://goreportcard.com/report/github.com/bowman2001/goldmark-supersubscript)

# goldmark-SuperSubscript

This Go module contains two extensions for the Markdown parser [Goldmark](https://github.com/yuin/goldmark) providing super- and subscripts.

## Syntax

Similar to [markdown-it](https://github.com/markdown-it/markdown-it) the new markup characters are:

- The circumflex `^` for superscript

- The tilde `~` for subscript

We need to place one before and one after the text segment like `H~2~O` or `x^2^`. 

**No whitespace** between the two surrounding markup characters is allowed. This way the common slip using TeX syntax like `x^2 + x^5` does not lead to messed up HTML. In case we definitely want to insert space we need to place a non-breaking space—either directly or as the HTML entity `&nbsp;`. 
