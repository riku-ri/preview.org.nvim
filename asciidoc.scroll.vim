let s:cmd = "asciidoctor -s -S secure -|cat ".'/'.join(split(expand('<sfile>'),'/')[0:-2],'/').'/github.css /dev/stdin'
let s:port_filename = getpid().'.'.'data.port'
let s:port = substitute(readfile(s:port_filename)[0], ":", "", "")

let s:socketsend = 'bin/socketsend -k move'
let s:send_output_to_port = '/'.join(split(expand("<sfile>"),'/')[0:-2],'/').'/'.s:socketsend.' '.s:port
execute '!echo '.'['.line('.').','.line('$').']'.'|'.s:send_output_to_port
