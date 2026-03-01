" Vim syntax file for xaligo DSL (.xal)
" Language:   xaligo DSL
" Maintainer: xaligo project
" URL:        https://github.com/ryo-arima/xaligo

if exists("b:current_syntax")
  finish
endif

" ── Comment ─────────────────────────────────────────────────────────────────
syntax region xalComment start="<!--" end="-->" contains=@NoSpell
highlight default link xalComment Comment

" ── Tag delimiters ──────────────────────────────────────────────────────────
syntax match xalDelimiter "[<>/]"
highlight default link xalDelimiter Delimiter

" ── Layout tags ─────────────────────────────────────────────────────────────
syntax keyword xalLayoutTag contained
      \ frame container row col
highlight default link xalLayoutTag Statement

" ── AWS group tags ──────────────────────────────────────────────────────────
syntax keyword xalGroupTag contained
      \ aws-cloud aws-cloud-alt region availability-zone security-group
      \ auto-scaling-group vpc private-subnet public-subnet
      \ server-contents corporate-data-center ec2-instance-contents
      \ spot-fleet aws-account aws-iot-greengrass-deployment
      \ aws-iot-greengrass elastic-beanstalk-container
      \ aws-step-functions-workflow generic-group
highlight default link xalGroupTag Type

" ── Special item/spacer/connection tags ────────────────────────────────────
syntax keyword xalItemTag contained item spacer connection
highlight default link xalItemTag Special

" ── Tag name region (matches the identifier after < or </) ──────────────────
" This region contains the tag name token so keywords fire inside it.
syntax region xalTagName
      \ start=+</\?\ze[a-z]+ end=+[>/[:space:]]+ keepend
      \ contains=xalLayoutTag,xalGroupTag,xalItemTag,xalTagOther
      \ transparent

" Fallback: any tag name not matched by the keyword groups
syntax match xalTagOther contained "[a-z][a-z0-9-]*"
highlight default link xalTagOther Identifier

" ── Known attributes ────────────────────────────────────────────────────────
syntax keyword xalAttr contained
      \ width height class layout gap title id src dst span row col
      \ visible border item-size arrowhead-size
      \ margin margin-top margin-right margin-bottom margin-left
highlight default link xalAttr PreProc

" ── Attribute values (double-quoted strings) ────────────────────────────────
syntax region xalString start=+"+ end=+"+ contained contains=xalSpacingClass
highlight default link xalString String

" ── Spacing class tokens inside class="..." ─────────────────────────────────
" Highlights pa-N, ma-N, px-N, py-N, mx-N, my-N, pt-N … ml-N
syntax match xalSpacingClass
      \ "\(pa\|ma\|px\|py\|mx\|my\|pt\|pr\|pb\|pl\|mt\|mr\|mb\|ml\)-[0-9]\+"
      \ contained
highlight default link xalSpacingClass Constant

" ── Attribute region (attr="value") inside a tag ────────────────────────────
syntax region xalAttrRegion
      \ start=+[a-z][a-z0-9-]*\s*=+ end=+"\|'+
      \ contained contains=xalAttr,xalString,xalDelimiter
      \ transparent keepend

" ── Full open/self-closing tag ───────────────────────────────────────────────
syntax region xalTag
      \ start="<[a-z]" end="/\?>"
      \ contains=xalComment,xalLayoutTag,xalGroupTag,xalItemTag,xalTagOther,
      \           xalAttr,xalString,xalDelimiter

" ── Closing tag ─────────────────────────────────────────────────────────────
syntax region xalCloseTag
      \ start="</" end=">"
      \ contains=xalLayoutTag,xalGroupTag,xalItemTag,xalTagOther,xalDelimiter

let b:current_syntax = "xal"
