import pyRserve
from http.server import BaseHTTPRequestHandler, HTTPServer

class testHTTPServer_RequestHandler(BaseHTTPRequestHandler):
  # GET
  def do_GET(self):
    conn = pyRserve.connect()
    conn.voidEval('two <- function(x){ x * 2}') #declare the function
    res = conn.eval('two(9)')

    self.send_response(200)
    self.send_header('Content-type','text/plain')
    self.end_headers()
    self.wfile.write(bytes('two(9) == %f' % res, "utf8"))
    return

def run():
  print('starting server...')

  server_address = ('0.0.0.0', 8080)
  httpd = HTTPServer(server_address, testHTTPServer_RequestHandler)
  print('running server...')
  httpd.serve_forever()

run()
