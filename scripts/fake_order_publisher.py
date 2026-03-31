import argparse
import json
import random
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone
from typing import Optional

FIRST_NAMES = [
    "Aarav", "Diya", "Ishaan", "Meera", "Rohan", "Kavya", "Arjun", "Nisha"
]
LAST_NAMES = ["Kumar", "Reddy", "Sharma", "Patel", "Iyer", "Verma", "Gupta", "Rao"]
PAYMENT_METHODS = ["upi", "cod", "card"]
PACKAGE_SIZES = ["small", "medium", "large"]

# Zone configurations with area mappings
ZONES = {
    1: {
        "name": "Zone-A",
        "city": "Bangalore",
        "state": "Karnataka",
        "areas": ["Whitefield", "Koramangala", "Indiranagar", "Bangalore City"]
    },
    2: {
        "name": "Zone-B",
        "city": "Bangalore",
        "state": "Karnataka",
        "areas": ["Koramangala", "Indiranagar", "Whitefield", "JP Nagar"]
    },
    3: {
        "name": "Zone-C",
        "city": "Mumbai",
        "state": "Maharashtra",
        "areas": ["Bandra", "Andheri", "Dadar", "Marine Drive"]
    },
    4: {
        "name": "Zone-D",
        "city": "Delhi",
        "state": "Delhi",
        "areas": ["Connaught Place", "Nehru Place", "Noida", "Gurgaon"]
    },
    5: {
        "name": "Tambaram",
        "city": "Chennai",
        "state": "Tamil Nadu",
        "areas": ["Tambaram", "Velachery", "Pallikaranai"]
    },
    6: {
        "name": "Selaiyur",
        "city": "Chennai",
        "state": "Tamil Nadu",
        "areas": ["Selaiyur", "Chromepet", "Pallikaranai"]
    },
    7: {
        "name": "Pallikaranai",
        "city": "Chennai",
        "state": "Tamil Nadu",
        "areas": ["Pallikaranai", "Velachery", "Tambaram"]
    },
    8: {
        "name": "Chromepet",
        "city": "Chennai",
        "state": "Tamil Nadu",
        "areas": ["Chromepet", "Selaiyur", "Tambaram"]
    },
    9: {
        "name": "Velachery",
        "city": "Chennai",
        "state": "Tamil Nadu",
        "areas": ["Velachery", "Tambaram", "Pallikaranai"]
    },
    10: {
        "name": "Medavakkam",
        "city": "Chennai",
        "state": "Tamil Nadu",
        "areas": ["Medavakkam", "Velachery", "Pallikaranai"]
    },
}

ZONE_BAND_FEE_INR = {
    "A": 25,
    "B": 40,
    "C": 65,
    "D": 85,
    "E": 120,
}

ZONE_ROUTE_PATTERNS = [
    ["A"],
    ["B", "A"],
    ["C", "B", "A"],
    ["D", "C", "B", "A"],
    ["E", "D", "C", "B", "A"],
]


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def get_area_for_zone(zone_id: int) -> str:
    """Get a random area for the given zone_id."""
    if zone_id in ZONES:
        return random.choice(ZONES[zone_id]["areas"])
    return "Unknown Area"


def random_zone_route_path() -> list[str]:
    return random.choice(ZONE_ROUTE_PATTERNS)


def compute_delivery_fee_inr(zone_route_path: list[str]) -> int:
    return sum(ZONE_BAND_FEE_INR.get(band, 0) for band in zone_route_path)


def random_order(idx: int, zone_id: Optional[int] = None) -> dict:
    """Generate a random order, optionally for a specific zone."""
    if zone_id is None:
        zone_id = random.choice(list(ZONES.keys()))
    
    zone_info = ZONES.get(zone_id, ZONES[1])
    pickup_area = get_area_for_zone(zone_id)
    drop_area = get_area_for_zone(zone_id)
    
    # Ensure pickup and drop are different
    while drop_area == pickup_area:
        drop_area = get_area_for_zone(zone_id)
    
    first = random.choice(FIRST_NAMES)
    last = random.choice(LAST_NAMES)
    customer_name = f"{first} {last}"
    payment_method = random.choice(PAYMENT_METHODS)
    amount = round(random.uniform(120, 1450), 2)
    package_size = random.choice(PACKAGE_SIZES)
    weight = {
        "small": round(random.uniform(0.2, 2.0), 2),
        "medium": round(random.uniform(2.1, 7.0), 2),
        "large": round(random.uniform(7.1, 20.0), 2),
    }[package_size]
    
    distance = {
        "small": round(random.uniform(1, 5), 1),
        "medium": round(random.uniform(2, 8), 1),
        "large": round(random.uniform(3, 10), 1),
    }[package_size]
    zone_route_path = random_zone_route_path()
    delivery_fee_inr = compute_delivery_fee_inr(zone_route_path)

    return {
        "order_id": f"ord-fake-{int(time.time())}-{idx}",
        "customer_name": customer_name,
        "customer_id": f"cust-{random.randint(10000, 99999)}",
        "customer_contact_number": f"+91{random.randint(9000000000, 9999999999)}",
        "address": f"{random.randint(1, 400)}, {pickup_area}, {zone_info['city']}",
        "payment_method": payment_method,
        "payment_amount": amount,
        "package_size": package_size,
        "package_weight_kg": weight,
        "pickup_area": pickup_area,
        "drop_area": drop_area,
        "distance_km": distance,
        "tip_inr": 0,
        "zone_route_path": zone_route_path,
        "delivery_fee_inr": delivery_fee_inr,
        "zone_id": zone_id,
        "status": "assigned",
        "assigned_at": now_iso(),
        "source": "fake-order-publisher",
    }


def post_json(url: str, payload: dict, timeout: int) -> tuple[int, str]:
    body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            text = resp.read().decode("utf-8", errors="replace")
            return resp.getcode(), text
    except urllib.error.HTTPError as http_err:
        text = http_err.read().decode("utf-8", errors="replace")
        return http_err.code, text
    except urllib.error.URLError as url_err:
        return 500, str(url_err)


def request_json(method: str, url: str, timeout: int, payload: dict | None = None, headers: dict | None = None) -> tuple[int, str]:
    data = None
    req_headers = {"Content-Type": "application/json"}
    if headers:
        req_headers.update(headers)
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")

    req = urllib.request.Request(url, data=data, headers=req_headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            text = resp.read().decode("utf-8", errors="replace")
            return resp.getcode(), text
    except urllib.error.HTTPError as http_err:
        text = http_err.read().decode("utf-8", errors="replace")
        return http_err.code, text
    except urllib.error.URLError as url_err:
        return 500, str(url_err)


def get_publisher_status(status_url: str, timeout: int, control_key: str) -> dict | None:
    """Fetch publisher control status from backend."""
    headers = {}
    if control_key:
        headers["X-Publisher-Key"] = control_key
    status, text = request_json("GET", status_url, timeout, headers=headers)
    if status >= 400:
        # Silently fail - might not be ready yet
        return None
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        return None


def fetch_available_orders(available_url: str, timeout: int, limit: int = 10, zone_id: Optional[int] = None) -> list[dict]:
    """Fetch available orders from backend API."""
    url = available_url
    if zone_id:
        url += f"?zone_id={zone_id}&limit={limit}"
    else:
        url += f"?limit={limit}"
    
    status, text = request_json("GET", url, timeout)
    if status >= 400:
        return []
    
    try:
        data = json.loads(text)
        return data.get("orders", [])
    except json.JSONDecodeError:
        return []


def fetch_deliveries(deliveries_url: str, timeout: int, limit: int = 50, zone_id: Optional[int] = None) -> list[dict]:
    """Fetch completed deliveries from backend API for tracking."""
    url = deliveries_url
    params = [f"limit={limit}", "status=delivered"]
    if zone_id:
        params.append(f"zone_id={zone_id}")
    
    url += "?" + "&".join(params)
    
    status, text = request_json("GET", url, timeout)
    if status >= 400:
        return []
    
    try:
        data = json.loads(text)
        return data.get("deliveries", [])
    except json.JSONDecodeError:
        return []


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Send fake orders to worker backend at a fixed interval, with zone-aware generation"
    )
    parser.add_argument(
        "--url",
        default="http://192.168.1.6:8001/api/v1/demo/orders/ingest",
        help="Ingest endpoint URL",
    )
    parser.add_argument(
        "--available-url",
        default="http://192.168.1.6:8001/api/v1/demo/orders/available",
        help="Fetch available orders endpoint URL",
    )
    parser.add_argument(
        "--deliveries-url",
        default="http://192.168.1.6:8001/api/v1/demo/deliveries",
        help="Fetch deliveries endpoint URL for tracking",
    )
    parser.add_argument(
        "--status-url",
        default="http://192.168.1.6:8001/api/v1/demo/orders/publisher/status",
        help="Backend control status URL (publishes only when active=true)",
    )
    parser.add_argument(
        "--control-key",
        default="",
        help="Optional X-Publisher-Key header value for control endpoints",
    )
    parser.add_argument(
        "--poll-seconds",
        type=int,
        default=5,
        help="How often to poll backend control status while waiting",
    )
    parser.add_argument(
        "--runs",
        type=int,
        default=2,
        help="How many publish cycles (default: 2)",
    )
    parser.add_argument(
        "--continuous",
        action="store_true",
        help="Run forever until interrupted (Ctrl+C)",
    )
    parser.add_argument(
        "--interval-seconds",
        type=int,
        default=60,
        help="Seconds between cycles (default: 60 => twice in 2 minutes)",
    )
    parser.add_argument(
        "--orders-per-run",
        type=int,
        default=1,
        help="Orders published per cycle",
    )
    parser.add_argument(
        "--timeout-seconds",
        type=int,
        default=15,
        help="HTTP timeout in seconds",
    )
    parser.add_argument(
        "--generate",
        action="store_true",
        help="Generate new random orders (default). If not set, fetches from backend.",
    )
    parser.add_argument(
        "--fetch",
        action="store_true",
        help="Fetch available orders from backend API instead of generating",
    )
    parser.add_argument(
        "--zone-id",
        type=int,
        help="Optional zone_id to filter/generate orders for specific zone",
    )
    parser.add_argument(
        "--show-deliveries",
        action="store_true",
        help="Periodically show completed deliveries from backend",
    )
    args = parser.parse_args()

    # Default to generate if neither flag is set
    if not args.fetch and not args.generate:
        args.generate = True

    print(f"Publishing to: {args.url}")
    print(f"Control status URL: {args.status_url}")
    print(f"Available orders URL: {args.available_url}")
    if args.show_deliveries:
        print(f"Deliveries URL: {args.deliveries_url}")
    
    mode = "fetch" if args.fetch else "generate"
    zone_filter = f" (zone_id={args.zone_id})" if args.zone_id else ""
    print(f"Mode: {mode}{zone_filter}")
    
    if args.continuous:
        print(
            f"Plan: continuous mode, {args.orders_per_run} order(s)/cycle, every {args.interval_seconds}s"
        )
        print(
            f"Behavior: publish only when backend initiates; lease expires after 5 minutes unless ack extends it"
        )
    else:
        print(
            f"Plan: {args.runs} runs, {args.orders_per_run} order(s)/run, every {args.interval_seconds}s"
        )

    run_no = 0
    current_session = ""
    try:
        while True:
            status_payload = get_publisher_status(args.status_url, args.timeout_seconds, args.control_key)
            if status_payload is None:
                time.sleep(max(1, args.poll_seconds))
                continue

            active = bool(status_payload.get("active", False))
            session_id = str(status_payload.get("session_id", "") or "")
            remaining_sec = int(status_payload.get("remaining_sec", 0) or 0)

            if not active:
                if current_session:
                    print(f"[control] lease expired for session {current_session}. Waiting for backend initiate/ack.")
                    current_session = ""
                time.sleep(max(1, args.poll_seconds))
                continue

            if session_id and session_id != current_session:
                current_session = session_id
                print(f"[control] active lease session={current_session}, remaining={remaining_sec}s")

            run_no += 1
            if args.continuous:
                print(f"\nRun {run_no} at {now_iso()} (session={current_session}, lease_remaining={remaining_sec}s)")
            else:
                print(f"\nRun {run_no}/{args.runs} at {now_iso()}")

            published_count = 0
            
            if args.fetch:
                # Fetch available orders from backend
                available = fetch_available_orders(args.available_url, args.timeout_seconds, limit=args.orders_per_run * 2, zone_id=args.zone_id)
                if available:
                    for i, order_data in enumerate(available[:args.orders_per_run]):
                        # Order already exists in backend, just noting it
                        print(f"  -> Available: {order_data.get('order_id')} from {order_data.get('zone_name')} (value={order_data.get('order_value')})")
                        published_count += 1
                else:
                    print(f"  -> No available orders to fetch")
            else:
                # Generate and publish new orders
                for i in range(args.orders_per_run):
                    payload = random_order(i + 1, zone_id=args.zone_id)
                    status, response_text = post_json(args.url, payload, args.timeout_seconds)
                    zone_name = ZONES.get(payload.get("zone_id", 1), {}).get("name", "Unknown")
                    print(f"  -> {payload['order_id']} (zone={zone_name}) status={status}")
                    if status >= 400:
                        print(f"     error={response_text}")
                    else:
                        published_count += 1
            
            # Show deliveries tracking if requested
            if args.show_deliveries and (run_no % 3 == 0 or run_no == 1):
                deliveries = fetch_deliveries(args.deliveries_url, args.timeout_seconds, limit=5, zone_id=args.zone_id)
                if deliveries:
                    print(f"  [Deliveries] Recent completed: {len(deliveries)} orders")
                    for d in deliveries[:3]:
                        print(f"    - {d.get('order_id')}: {d.get('worker_name')} in {d.get('zone_name')}")

            if not args.continuous and run_no >= args.runs:
                break

            time.sleep(args.interval_seconds)
    except KeyboardInterrupt:
        print("\nStopped by user.")

    print("\nDone.")


if __name__ == "__main__":
    main()
