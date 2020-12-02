// Variables in client-server enviroment
//   url     : request url, setted with context.WithValue("request-url", "....")
//   ua      : client user-agent, setted with context.WithValue("user-agent", "...")
//   ip      : client_ip, setted with context.WithValue("client-ip", "...")
//   get.xxx : query value from url. e.g. url: xxx/a=1&b=1,2,3&c={"foo":[{"bar":1}]}
//              get.a => "1"
//              get.b => "1,2,3"
//              get.b[0] => "1"  // list string or array element access
//              get.b[2] => "3"
//              get.c{foo.0.bar} => "1"   // json string access, pass json path
//
package request
