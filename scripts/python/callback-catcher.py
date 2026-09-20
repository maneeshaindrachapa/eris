from http.server import BaseHTTPRequestHandler, HTTPServer

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/html")
        self.end_headers()
        self.wfile.write(f"""
            <h2>Callback received</h2>
            <p>Full URL: {self.path}</p>
            <p>Copy the <code>code</code> value above and use it in step 4.</p>
        """.encode())

HTTPServer(("localhost", 3000), Handler).serve_forever()