"g:preview variables
"g:preview variables
if !exists('g:preview_root')
	let g:preview_root = '/'.join(split(expand("<sfile>"),'/')[0:-2],'/')
endif

if !exists('g:preview_tmp')
	let g:preview_tmp = '.'
endif

if !exists('g:preview_browser')
	let g:preview_browser = 'curl'
elseif(g:preview_browser == "")
	let g:preview_browser = 'curl'
endif

let g:preview_websocket_program = g:preview_root.'/'.'websocket'

if !exists('g:preview_timer_mode')
	let g:preview_timer_mode = 0
endif
if !exists('g:preview_timer_gap')
	let g:preview_timer_gap = 500
endif
"g:preview variables
"g:preview variables

"vim compat
function! s:jobstop(job)
	if exists('*jobstop')
		return jobstop(a:job)
	elseif exists('*job_stop')
		return job_stop(a:job)
	else
		echoerr "This vi-like editor can not stop a job"
		finish
	endif
endfunction

function! s:jobstart(cmd)
	if exists('*jobstart')
		return jobstart(a:cmd)
	elseif exists('*job_start')
		return job_start(a:cmd)
	else
		echoerr "This vi-like editor can not start a job"
		finish
	endif
endfunction

function! s:chansend(job, data)
	if exists('*chansend')
		return chansend(a:job, a:data)
	elseif exists('*ch_sendraw')
		return ch_sendraw(a:job, a:data)
	else
		echoerr "This vi-like editor can send data to STDIN of a job"
		finish
	endif
endfunction
"vim compat

"process
function! s:preview_stop()
	if exists('b:preview_websocketid')
		call <SID>jobstop(b:preview_websocketid)
	endif
	if exists('b:preview_browserid')
		call <SID>jobstop(b:preview_browserid)
	endif
endfunction

let s:port_filename = g:preview_tmp.'/'.getpid().'.'.'data.port'

let b:preview_websocketid = <SID>jobstart([g:preview_websocket_program , g:preview_tmp , getpid()])

"TODO : check job start for vim
if has('nvim')
if b:preview_websocketid <= 0
	echoerr "file `" g:preview_websocket_program "` can not be executed or can not setup HTTP server"
	finish
endif
endif

"Server need preview root directory to create unique file
"Server need vim pid to create unique file
"Server need vim buffer id to create unique file
"Server need '\xff' to ensure sending is finished
"echo <SID>chansend(b:preview_websocketid, g:preview_tmp."\xff") "byte has been send to HTTP server process for" "tmpdir"
"echo <SID>chansend(b:preview_websocketid, getpid()."\xff")      "byte has been send to HTTP server process for" "pid"
"echo <SID>chansend(b:preview_websocketid, bufnr('%')."\xff")    "byte has been send to HTTP server process for" "bufnr"

echo s:port_filename
echo "Wait for HTTP Server startup"
while(!filereadable(s:port_filename))
endwhile
""Wait for set up server
echo "HTTP Server startup has finished"
echo "please wait for browser opening"

let s:preview_websocket_html = g:preview_tmp.'/'.getpid().'.'.'websocket.html'
if g:preview_browser == "chromium" || g:preview_browser == "firefox"
	let b:preview_browserid = <SID>jobstart([g:preview_browser, "--new-window", s:preview_websocket_html])
else
	let b:preview_browserid = <SID>jobstart([g:preview_browser, s:preview_websocket_html])
endif

function! s:preview_timer_call(timer)
	silent execute 'so '.g:preview_root.'/'.&ft.'.body.vim'
endfunction

if(filereadable(g:preview_root.'/'.&ft.'.body.vim'))
	if !g:preview_timer_mode
		execute 'so '.g:preview_root.'/'.&ft.'.body.vim'
		execute "augroup preview".&ft
			autocmd!
			autocmd InsertLeave,TextChanged,TextChangedI <buffer> silent execute 'so '.g:preview_root.'/'.&ft.'.body.vim'
			"autocmd CursorMoved <buffer> silent execute 'so '.g:preview_root.'/'.&ft.'.scroll.vim'
			autocmd BufUnload <buffer> call <SID>preview_stop()
		execute "augroup END"
	else
		let g:tex_compile_timer_id = timer_start(g:preview_timer_gap, 's:preview_timer_call', {'repeat': -1})
		if(g:tex_compile_timer_id == -1)
			echoerr "Can not run tex timer"
		endif
	endif
else
	echoerr "file `" g:preview_root.'/'.&ft.'.body.vim' "` can not be read"
	call <SID>preview_stop()
	finish
endif
