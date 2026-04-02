import time
import random
import json
import urllib.request
import urllib.error
import argparse
import os
from datetime import datetime, timezone

# Load environment variables
env_file = os.path.join(os.path.dirname(__file__), "../.env")
if os.path.exists(env_file):
    with open(env_file, "r") as f:
        for line in f:
            if "=" in line and not line.startswith("#"):
                parts = line.strip().split("=", 1)
                if len(parts) == 2:
                    os.environ[parts[0]] = parts[1]

# Prioritize the Platform Gateway port (8003) for Part 2 metrics
HOST_IP = os.getenv("HOST_IP", "localhost")
DEFAULT_BASE_URL = f"http://{HOST_IP}:8003"

def post_json(url, data):
    body = json.dumps(data).encode("utf-8")
    req = urllib.request.Request(
        url, 
        data=body, 
        headers={"Content-Type": "application/json"},
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            return resp.getcode(), resp.read().decode("utf-8")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8")
    except Exception as e:
        return 500, str(e)

def main():
    parser = argparse.ArgumentParser(description="Live Order Stream Simulator for InDel Part-2")
    parser.add_argument("--zone", type=int, default=1, help="Zone ID to simulate")
    parser.add_argument("--port", type=int, default=8003, help="Platform Gateway port")
    args = parser.parse_args()

    target_url = f"http://{HOST_IP}:{args.port}/api/v1/platform/webhooks/order/completed"

    print(f"Starting InDel Auto-Stream Simulator")
    print(f"Target: {target_url}")
    print(f"Zone: {args.zone}")
    print(f"Behavior: Healthy flow for 15s, then automatic volume drop.")
    print(f"------------------------------------------")

    start_time = time.time()
    
    def check_for_reset(current_start_time):
        health_url = f"http://{HOST_IP}:{args.port}/api/v1/platform/zones/health"
        try:
            req = urllib.request.Request(health_url, method="GET")
            with urllib.request.urlopen(req, timeout=2) as resp:
                data = json.loads(resp.read().decode("utf-8"))
                for zone in data.get("data", {}).get("data", []):
                    if zone.get("zone_id") == args.zone:
                        last_reset = zone.get("last_reset_at", 0)
                        if last_reset > current_start_time:
                            print(f"\n🔄 [SYNC] Backend reset detected! Restarting simulation pulse...")
                            return time.time()
        except Exception:
            pass
        return current_start_time

    try:
        while True:
            # Sync check
            start_time = check_for_reset(start_time)
            
            order_id = f"ord-fake-{int(time.time())}-{random.randint(1000, 9999)}"
            payload = {
                "order_id": order_id,
                "amount": float(random.randint(50, 200)),
                "zone_id": args.zone,
                "completed_at": datetime.now(timezone.utc).isoformat()
            }

            status_code, response = post_json(target_url, payload)
            
            elapsed = time.time() - start_time
            is_healthy = elapsed < 15
            
            icon = "✅" if status_code == 200 else "❌"
            mode_tag = "[NORMAL]" if is_healthy else "[DROP]"
            
            print(f"[{datetime.now().strftime('%H:%M:%S')}] {mode_tag} {icon} {order_id}")
            
            if status_code >= 400:
                print(f"   ⚠️ Error: {response}")

            # Auto-transition logic
            if is_healthy:
                time.sleep(random.uniform(0.8, 1.5)) # ~1 order per second
            else:
                # Polling for reset during the long sleep
                for _ in range(15):
                    time.sleep(1)
                    new_start = check_for_reset(start_time)
                    if new_start != start_time:
                        start_time = new_start
                        break # Break wait to send new order immediately


    except KeyboardInterrupt:
        print("\n👋 Simulator stopped.")

if __name__ == "__main__":
    main()
