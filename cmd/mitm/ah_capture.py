import json
from mitmproxy import http

endpoints = {}

class AHCapture:
    def response(self, flow: http.HTTPFlow):
        if "api.ah.nl" not in flow.request.host:
            return

        method = flow.request.method
        path = flow.request.path.split("?")[0]
        key = f"{method} {path}"

        entry = {
            "url": flow.request.url,
            "status": flow.response.status_code,
            "request_headers": dict(flow.request.headers),
        }

        # Capture request body
        if flow.request.content:
            try:
                entry["request_body"] = json.loads(flow.request.content)
            except:
                entry["request_body"] = flow.request.content.decode("utf-8", errors="ignore")

        # Capture response body
        if flow.response.content:
            try:
                entry["response_body"] = json.loads(flow.response.content)
            except:
                pass

        endpoints[key] = entry

        # Log GraphQL operations
        if "/graphql" in path and "request_body" in entry:
            if isinstance(entry["request_body"], dict):
                query = entry["request_body"].get("query", "")
                if "query " in query or "mutation " in query:
                    op = query.split("(")[0].split("{")[0].strip()
                    print(f"GraphQL: {op}")

        print(f"{method} {path} -> {flow.response.status_code}")

    def done(self):
        with open("ah_endpoints.json", "w") as f:
            json.dump(endpoints, f, indent=2, default=str)
        print(f"\nSaved {len(endpoints)} endpoints to ah_endpoints.json")

addons = [AHCapture()]
