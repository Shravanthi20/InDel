
import argparse
import json
import random
import time
import urllib.error
import urllib.request
import urllib.parse
from datetime import datetime, timezone
from typing import Optional
import os


def resolve_default_host() -> str:
    """Resolve API host from env var or project .env, defaulting to localhost."""
    host = os.getenv("HOST_IP") or os.getenv("INDEL_HOST_IP")
    if host:
        return host

    env_path = os.path.join(os.path.dirname(__file__), "..", ".env")
    try:
        with open(env_path, encoding="utf-8") as env_file:
            for line in env_file:
                line = line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, value = line.split("=", 1)
                if key.strip() == "HOST_IP" and value.strip():
                    return value.strip()
    except OSError:
        pass

    return "127.0.0.1"


DEFAULT_HOST = resolve_default_host()

FIRST_NAMES = [
    "Aarav", "Diya", "Ishaan", "Meera", "Rohan", "Kavya", "Arjun", "Nisha"
]
LAST_NAMES = ["Kumar", "Reddy", "Sharma", "Patel", "Iyer", "Verma", "Gupta", "Rao"]
PAYMENT_METHODS = ["upi", "cod", "card"]
PACKAGE_SIZES = ["small", "medium", "large"]
ZONES = {}
MAX_ZONE_NAMES_PER_LEVEL = 10
ORDERS_PER_ZONE_NAME = 5


# Load zone pairs from zone_a.json (same city), zone_b.json (intra-state) and zone_c.json (inter-state)
def load_zone_pairs():
    base_dir = os.path.dirname(__file__)
    with open(os.path.join(base_dir, '../zone_a.json'), encoding='utf-8') as f:
        zone_a = json.load(f)
    with open(os.path.join(base_dir, '../zone_b.json'), encoding='utf-8') as f:
        zone_b = json.load(f)
    with open(os.path.join(base_dir, '../zone_c.json'), encoding='utf-8') as f:
        zone_c = json.load(f)
    return zone_a, zone_b, zone_c


def limit_zone_pairs(zone_a_pairs, zone_b_pairs, zone_c_pairs, limit: int = MAX_ZONE_NAMES_PER_LEVEL):
    return zone_a_pairs[:limit], zone_b_pairs[:limit], zone_c_pairs[:limit]

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


def print_zones(city: Optional[str] = None) -> None:
    """Print zone metadata if it is available."""
    print("Available Zones:")
    if not ZONES:
        print("  (zone metadata not loaded in this build)")
        return
    for zid, zinfo in ZONES.items():
        if city and zinfo.get("city", "").lower() != city.lower():
            continue
        print(f"  Zone ID: {zid}")
        print(f"    Name: {zinfo.get('name', 'Unknown')}")
        print(f"    City: {zinfo.get('city', 'Unknown')}")
        print(f"    State: {zinfo.get('state', 'Unknown')}")
        print(f"    Areas: {', '.join(zinfo.get('areas', []))}")


def random_zone_route_path() -> list[str]:
    return random.choice(ZONE_ROUTE_PATTERNS)


def compute_delivery_fee_inr(zone_route_path: list[str]) -> int:
    return sum(ZONE_BAND_FEE_INR.get(band, 0) for band in zone_route_path)



# Generate a random order from a zone pair (from/to cities)
def random_order_from_pair(idx: int, pair: dict, zone_type: str, zone_id: int = 1) -> dict:
    first = random.choice(FIRST_NAMES)
    last = random.choice(LAST_NAMES)
    customer_name = f"{first} {last}"
    payment_method = random.choice(PAYMENT_METHODS)
    amount = round(random.uniform(120, 1450), 2)
    package_size = random.choice(PACKAGE_SIZES)
    weight = {
        "small": round(random.uniform(0.05, 1.0), 2),
        "medium": round(random.uniform(1.0, 3.0), 2),
        "large": round(random.uniform(3.0, 5.0), 2),
    }[package_size]
    weight = min(max(weight, 0.05), 5.0)
    distance = pair.get("distance_km", round(random.uniform(5, 1000), 1))
    # Use from/to city/state/lat/lon
    from_city = pair.get("from")
    to_city = pair.get("to")
    from_state = pair.get("from_state") or pair.get("state")
    to_state = pair.get("to_state") or pair.get("state")
    from_lat = pair.get("from_lat")
    from_lon = pair.get("from_lon")
    to_lat = pair.get("to_lat")
    to_lon = pair.get("to_lon")
    delivery_fee_inr = int(distance * 2) if zone_type in ("zone_c", "inter-state") else int(distance * 1.2)
    zone_route_path = random_zone_route_path()
    eligibility = determine_zone_and_vehicle(from_city, to_city, CITY_STATE_LOOKUP)
    return {
        "order_id": f"ord-fake-{int(time.time())}-{idx}",
        "customer_name": customer_name,
        "customer_id": f"cust-{random.randint(10000, 99999)}",
        "customer_contact_number": f"+91{random.randint(9000000000, 9999999999)}",
        "address": f"{random.randint(1, 400)}, {from_city}, {from_state}",
        "payment_method": payment_method,
        "payment_amount": amount,
        "order_value": amount,
        "package_size": package_size,
        "package_weight_kg": weight,
        "pickup_area": from_city,
        "drop_area": to_city,
        "distance_km": distance,
        "tip_inr": 0,
        "zone_path": [from_city, to_city],
        "zone_route_path": zone_route_path,
        "delivery_fee_inr": delivery_fee_inr,
        "zone_id": zone_id,
        "zone_level": "A" if zone_type == "zone_a" else "B" if zone_type == "zone_b" else "C" if zone_type == "zone_c" else "",
        "source_zone_file": "zone_a.json" if zone_type == "zone_a" else "zone_b.json" if zone_type == "zone_b" else "zone_c.json" if zone_type == "zone_c" else "",
        "status": "assigned",
        "assigned_at": now_iso(),
        "source": "fake-order-publisher",
        # --- Eligibility fields ---
        "zone_type": eligibility["zone_type"],
        "required_vehicle_type": eligibility["required_vehicle_type"],
        "needs_hub_transfer": eligibility["needs_hub_transfer"],
        # --- New fields ---
        "from_city": from_city,
        "to_city": to_city,
        "from_state": from_state,
        "to_state": to_state,
        "from_lat": from_lat,
        "from_lon": from_lon,
        "to_lat": to_lat,
        "to_lon": to_lon,
        "vehicle_type": eligibility["required_vehicle_type"],
        "vehicle_capacity": 15 if package_size == "small" else 30 if package_size == "medium" else 50,
        "allowed_zones": f"{from_city},{to_city}",
        # --- Current node ---
        "current_node": from_city,
    }


def build_zone_a_pair_dicts(zone_a_cities: list[str]) -> list[dict]:
    pair_dicts = []
    for city in zone_a_cities:
        state = CITY_STATE_LOOKUP.get(city, "Unknown")
        pair_dicts.append(
            {
                "from": city,
                "to": city,
                "from_state": state,
                "to_state": state,
                "distance_km": 1.0,
                "from_lat": 0,
                "from_lon": 0,
                "to_lat": 0,
                "to_lon": 0,
            }
        )
    return pair_dicts


def build_all_pairs(zone_a_pairs: list[str], zone_b_pairs: list[dict], zone_c_pairs: list[dict]) -> list[tuple[dict, str]]:
    zone_a_pair_dicts = build_zone_a_pair_dicts(zone_a_pairs)
    return (
        [(pair, "zone_a") for pair in zone_a_pair_dicts]
        + [(pair, "zone_b") for pair in zone_b_pairs]
        + [(pair, "zone_c") for pair in zone_c_pairs]
    )


def publish_one_cycle(args, all_pairs: list[tuple[dict, str]], cycle_index: int, global_order_index: int) -> tuple[int, int, int]:
    published_count = 0
    failed_count = 0
    order_index = global_order_index

    print(f"\n=== Publish Cycle {cycle_index} ===")
    print(f"Routes considered: {len(all_pairs)} | Orders per route: {args.orders_per_run}")

    for pair_index, (pair, zone_type) in enumerate(all_pairs):
        for route_order_index in range(args.orders_per_run):
            order_index += 1
            payload = random_order_from_pair(order_index, pair, zone_type, args.zone_id or 1)
            status, response_text = post_json(args.url, payload, args.timeout_seconds, retries=2)
            print(
                f"  -> {payload['order_id']} {payload['from_city']}->{payload['to_city']} "
                f"[{payload.get('source_zone_file', zone_type)}] status={status}"
            )

            if status >= 400:
                print(f"     error={response_text}")
                failed_count += 1
            else:
                published_count += 1

            is_last_order = pair_index == len(all_pairs) - 1 and route_order_index == args.orders_per_run - 1
            if not is_last_order:
                time.sleep(0.005)

    print(f"Cycle {cycle_index} done. published={published_count}, failed={failed_count}")
    return published_count, failed_count, order_index
def reset_backend_orders(reset_url: str, timeout: int):
    print(f"Resetting backend orders via {reset_url} ...")
    try:
        status, text = request_json("POST", reset_url, timeout)
        print(f"Reset status={status}: {text}")
    except Exception as e:
        print(f"Reset failed: {e}")



def post_json(url: str, payload: dict, timeout: int, retries: int = 3) -> tuple[int, str]:
    body = json.dumps(payload).encode("utf-8")
    last_error = None
    
    for attempt in range(retries):
        try:
            req = urllib.request.Request(
                url,
                data=body,
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                text = resp.read().decode("utf-8", errors="replace")
                return resp.getcode(), text
        except urllib.error.HTTPError as http_err:
            text = http_err.read().decode("utf-8", errors="replace")
            return http_err.code, text
        except (urllib.error.URLError, TimeoutError) as err:
            last_error = err
            if attempt < retries - 1:
                # Exponential backoff: 0.5s, 1s, 2s
                wait_time = 0.5 * (2 ** attempt)
                # Skip sleep on first attempt to avoid delays
                if attempt > 0:
                    time.sleep(wait_time)
                continue
            else:
                return 500, f"Timeout after {retries} retries: {str(err)}"
    
    return 500, str(last_error)


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


def probe_endpoint(url: str, timeout: int) -> tuple[bool, str]:
    """Return whether an endpoint is reachable at the network level.

    A 404/405 still counts as reachable, because the TCP connection succeeded.
    """
    req = urllib.request.Request(url, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return True, f"HTTP {resp.getcode()}"
    except urllib.error.HTTPError as http_err:
        return True, f"HTTP {http_err.code}"
    except urllib.error.URLError as url_err:
        return False, str(url_err)


def suggest_host_fix(url: str) -> str:
    parsed = urllib.parse.urlparse(url)
    host = parsed.hostname or ""
    if host.startswith("192.168.") or host.startswith("10.") or host.startswith("172."):
        return (
            f"URL host '{host}' looks like a private LAN IP. "
            "Ensure HOST_IP in .env matches this same LAN IP and containers are restarted."
        )
    return (
        "Verify the URL host/port and that docker compose services are up. "
        f"Example: http://{DEFAULT_HOST}:8001/api/v1/demo/reset"
    )


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


# --- City-State Lookup Utility ---
import os
import csv
from typing import Dict

def load_city_state_lookup(csv_path: str) -> Dict[str, str]:
    """
    Loads a mapping from area name (city in CSV, stripped) to city name (state in CSV).
    Args:
        csv_path: Path to the Indian Cities Geo Data CSV file.
    Returns:
        Dictionary mapping area name (city, e.g. 'Port Blair') to city name (state).
    """
    lookup = {}
    with open(csv_path, newline='', encoding='utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            area = row['Location'].replace(' Latitude and Longitude', '').strip()
            city = row['State'].strip()
            lookup[area] = city
    return lookup

CITY_STATE_CSV = os.path.join(os.path.dirname(__file__), '../Indian Cities Geo Data.csv')
CITY_STATE_LOOKUP = load_city_state_lookup(CITY_STATE_CSV)


# --- Zone Rule Engine ---
def determine_zone_and_vehicle(source_city: str, dest_city: str, city_state_lookup: dict) -> dict:
    """
    Determines the zone type and required vehicle for an order.
    Args:
        source_city: Name of the source city.
        dest_city: Name of the destination city.
        city_state_lookup: Dict mapping city name to state name.
    Returns:
        Dict with zone_type, required_vehicle_type, needs_hub_transfer.
    """
    source_state = city_state_lookup.get(source_city)
    dest_state = city_state_lookup.get(dest_city)
    if not source_state or not dest_state:
        return {
            'zone_type': 'unknown',
            'required_vehicle_type': 'unknown',
            'needs_hub_transfer': False
        }
    if source_state == dest_state:
        return {
            'zone_type': 'intra-zone',
            'required_vehicle_type': 'bike/small van',
            'needs_hub_transfer': False
        }
    else:
        # For now, assume all inter-state orders are possible directly
        return {
            'zone_type': 'inter-state',
            'required_vehicle_type': 'van/truck',
            'needs_hub_transfer': False  # Set to True if you want to force hub transfer for some cases
        }

# Example usage:
# result = determine_zone_and_vehicle('Chennai', 'Bangalore', CITY_STATE_LOOKUP)
# print(result)




    """Print available zones, optionally filtered by city name."""
    print("Available Zones:")
    for zid, zinfo in ZONES.items():
        if city and zinfo["city"].lower() != city.lower():
            continue
        print(f"  Zone ID: {zid}")
        print(f"    Name: {zinfo['name']}")
        print(f"    City: {zinfo['city']}")
        print(f"    State: {zinfo['state']}")
        print(f"    Areas: {', '.join(zinfo['areas'])}")

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Send fake orders to worker backend at a fixed interval, with zone-aware generation"
    )
    parser.add_argument(
        "--url",
        default=f"http://{DEFAULT_HOST}:8001/api/v1/demo/orders/ingest",
        help="Ingest endpoint URL",
    )
    parser.add_argument(
        "--available-url",
        default=f"http://{DEFAULT_HOST}:8001/api/v1/demo/orders/available",
        help="Fetch available orders endpoint URL",
    )
    parser.add_argument(
        "--deliveries-url",
        default=f"http://{DEFAULT_HOST}:8001/api/v1/demo/deliveries",
        help="Fetch deliveries endpoint URL for tracking",
    )
    parser.add_argument(
        "--status-url",
        default=f"http://{DEFAULT_HOST}:8001/api/v1/demo/orders/publisher/status",
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
        default=5,
        help="Orders published per route pair before moving to the next pair",
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
    parser.add_argument(
        "--list-zones",
        action="store_true",
        help="List available zones and exit",
    )
    parser.add_argument(
        "--city",
        type=str,
        help="Filter zones by city name (used with --list-zones)",
    )

    parser.add_argument(
        "--reset-url",
        default=f"http://{DEFAULT_HOST}:8001/api/v1/demo/reset",
        help="Backend reset endpoint URL (POST)",
    )

    args = parser.parse_args()

    if args.list_zones:
        print_zones(args.city)
        return


    # Default to generate if neither flag is set
    if not args.fetch and not args.generate:
        args.generate = True

    # Fail fast when the target host is unreachable to avoid long noisy retries.
    reachable, detail = probe_endpoint(args.reset_url, args.timeout_seconds)
    if not reachable:
        print(f"Preflight failed for reset endpoint: {args.reset_url}")
        print(f"Reason: {detail}")
        print(suggest_host_fix(args.reset_url))
        return

    reachable, detail = probe_endpoint(args.url, args.timeout_seconds)
    if not reachable:
        print(f"Preflight failed for ingest endpoint: {args.url}")
        print(f"Reason: {detail}")
        print(suggest_host_fix(args.url))
        return

    # Always reset backend before publishing
    reset_backend_orders(args.reset_url, args.timeout_seconds)

    # Load zone pairs
    zone_a_pairs, zone_b_pairs, zone_c_pairs = load_zone_pairs()
    zone_a_pairs, zone_b_pairs, zone_c_pairs = limit_zone_pairs(zone_a_pairs, zone_b_pairs, zone_c_pairs)

    print(f"Publishing to: {args.url}")
    print(f"Available orders URL: {args.available_url}")
    print(
        f"Zone sources: zone_a.json / zone_b.json / zone_c.json "
        f"(first {MAX_ZONE_NAMES_PER_LEVEL} route entries from each)."
    )
    print(f"Mode: {'continuous' if args.continuous else f'{args.runs} run(s)'} | interval={args.interval_seconds}s | orders-per-run={args.orders_per_run}")

    all_pairs = build_all_pairs(zone_a_pairs, zone_b_pairs, zone_c_pairs)
    if not all_pairs:
        print("No zone pairs loaded; aborting publish.")
        return

    total_published = 0
    total_failed = 0
    cycle_count = 0
    order_index = 0

    try:
        while args.continuous or cycle_count < args.runs:
            cycle_count += 1

            if args.fetch:
                available = fetch_available_orders(args.available_url, args.timeout_seconds, limit=20, zone_id=args.zone_id)
                print(f"\n=== Fetch Cycle {cycle_count} ===")
                print(f"Fetched {len(available)} currently available orders")
            else:
                published, failed, order_index = publish_one_cycle(args, all_pairs, cycle_count, order_index)
                total_published += published
                total_failed += failed

            if args.show_deliveries:
                deliveries = fetch_deliveries(args.deliveries_url, args.timeout_seconds, limit=10, zone_id=args.zone_id)
                print(f"Recent delivered orders: {len(deliveries)}")

            should_continue = args.continuous or cycle_count < args.runs
            if should_continue:
                print(f"Sleeping {args.interval_seconds}s before next cycle...")
                time.sleep(max(1, args.interval_seconds))
    except KeyboardInterrupt:
        print("\nStopped by user.")

    print(f"\nDone. total_published={total_published}, total_failed={total_failed}, cycles={cycle_count}")

if __name__ == "__main__":
    main()
