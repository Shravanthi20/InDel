"""
Route Optimizer with Depot Location Support
- User can specify a depot city (must exist in city coords)
- All vehicles start at the depot
- Orders are indexed after the depot in the distance matrix
"""

import sys
import os
import random
from typing import List, Dict
from ortools.constraint_solver import routing_enums_pb2
from ortools.constraint_solver import pywrapcp
import math
import matplotlib.pyplot as plt

sys.path.append(os.path.dirname(__file__))
from fake_order_publisher import random_order, CITY_STATE_LOOKUP, ZONES

def haversine(lat1, lon1, lat2, lon2):
    R = 6371
    phi1, phi2 = math.radians(lat1), math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lon2 - lon1)
    a = math.sin(dphi/2)**2 + math.cos(phi1)*math.cos(phi2)*math.sin(dlambda/2)**2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))

def load_city_coords(csv_path: str) -> Dict[str, tuple]:
    import csv
    coords = {}
    with open(csv_path, newline='', encoding='utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            city = row['Location'].replace(' Latitude and Longitude', '').strip()
            lat = float(row['Latitude'])
            lon = float(row['Longitude'])
            coords[city] = (lat, lon)
    return coords

CITY_COORDS = load_city_coords(os.path.join(os.path.dirname(__file__), '../Indian Cities Geo Data.csv'))

def build_distance_matrix_with_depot(depot_city, orders):
    n = len(orders) + 1  # +1 for depot
    matrix = [[0]*n for _ in range(n)]
    depot_coords = CITY_COORDS[depot_city]

    # Helper: map area to parent city using ZONES config
    def area_to_city(area):
        for z in ZONES.values():
            if area in z['areas']:
                return z['city']
        return area  # fallback

    order_coords = []
    for o in orders:
        city = area_to_city(o['pickup_area'])
        if city in CITY_COORDS:
            order_coords.append(CITY_COORDS[city])
        else:
            # fallback: use depot coords if city not found
            order_coords.append(depot_coords)

    # Depot to orders
    for j in range(1, n):
        matrix[0][j] = int(haversine(*depot_coords, *order_coords[j-1]) * 1000)
    # Orders to depot
    for i in range(1, n):
        matrix[i][0] = int(haversine(*order_coords[i-1], *depot_coords) * 1000)
    # Orders to orders
    for i in range(1, n):
        for j in range(1, n):
            if i == j:
                matrix[i][j] = 0
            else:
                matrix[i][j] = int(haversine(*order_coords[i-1], *order_coords[j-1]) * 1000)
    return matrix

def plot_routes_with_depot(depot_city, orders, routes, city_coords):
    plt.figure(figsize=(8, 6))
    colors = ['b', 'g', 'r', 'c', 'm', 'y', 'k']
    depot_lat, depot_lon = city_coords[depot_city]
    plt.scatter([depot_lon], [depot_lat], marker='*', color='orange', s=200, label='Depot')
    for v, route in enumerate(routes):
        color = colors[v % len(colors)]
        xs, ys = [depot_lon], [depot_lat]
        for idx in route:
            city = orders[idx]['pickup_area']
            if city in city_coords:
                lat, lon = city_coords[city]
                xs.append(lon)
                ys.append(lat)
        plt.plot(xs, ys, marker='o', color=color, label=f'Vehicle {v}')
    plt.xlabel('Longitude')
    plt.ylabel('Latitude')
    plt.title(f'Optimized Delivery Routes (Depot: {depot_city})')
    plt.legend()
    plt.show()


def solve_vrp_with_depot(orders: List[dict], depot_city: str, vehicle_configs=None, visualize: bool = True):
    if depot_city not in CITY_COORDS:
        print(f"Depot city '{depot_city}' not found in city coordinates!")
        return
    if vehicle_configs is None:
        # Default: 2 vehicles, one bike, one van/truck
        vehicle_configs = [
            {"type": "bike", "capacity": 15, "allowed_zones": None, "color": "b"},
            {"type": "van/truck", "capacity": 30, "allowed_zones": None, "color": "g"},
        ]
    num_vehicles = len(vehicle_configs)
    dist_matrix = build_distance_matrix_with_depot(depot_city, orders)
    manager = pywrapcp.RoutingIndexManager(len(dist_matrix), num_vehicles, 0)
    routing = pywrapcp.RoutingModel(manager)

    def distance_callback(from_index, to_index):
        from_node = manager.IndexToNode(from_index)
        to_node = manager.IndexToNode(to_index)
        return dist_matrix[from_node][to_node]

    transit_callback_index = routing.RegisterTransitCallback(distance_callback)
    routing.SetArcCostEvaluatorOfAllVehicles(transit_callback_index)

    # Capacity constraint
    demands = [0] + [int(order['package_weight_kg']) for order in orders]
    vehicle_capacities = [v["capacity"] for v in vehicle_configs]
    def demand_callback(from_index):
        from_node = manager.IndexToNode(from_index)
        return demands[from_node]
    demand_callback_index = routing.RegisterUnaryTransitCallback(demand_callback)
    routing.AddDimensionWithVehicleCapacity(
        demand_callback_index, 0, vehicle_capacities, True, 'Capacity')

    # Time window constraint (random for demo)
    time_windows = [(0, 10000)] * (len(orders) + 1)
    for i in range(1, len(orders) + 1):
        start = random.randint(0, 5000)
        end = start + random.randint(1000, 5000)
        time_windows[i] = (start, end)
    def time_callback(from_index, to_index):
        return dist_matrix[manager.IndexToNode(from_index)][manager.IndexToNode(to_index)] // 10
    time_callback_index = routing.RegisterTransitCallback(time_callback)
    routing.AddDimension(
        time_callback_index, 1000, 100000, False, 'Time')
    time_dimension = routing.GetDimensionOrDie('Time')
    for idx, window in enumerate(time_windows):
        index = manager.NodeToIndex(idx)
        time_dimension.CumulVar(index).SetRange(*window)

    # Vehicle type and allowed zones constraint
    order_vehicle_types = ['depot'] + [order.get('required_vehicle_type', 'bike') for order in orders]
    order_zones = ['depot'] + [order.get('zone_id', None) for order in orders]
    for idx, (req_type, req_zone) in enumerate(zip(order_vehicle_types, order_zones)):
        if idx == 0:
            continue
        allowed_vehicles = []
        for v_idx, v in enumerate(vehicle_configs):
            type_match = (v["type"] == req_type) or (req_type in v["type"]) or (v["type"] in req_type)
            zone_match = (v["allowed_zones"] is None) or (req_zone in (v["allowed_zones"] or []))
            if type_match and zone_match:
                allowed_vehicles.append(v_idx)
        if allowed_vehicles:
            routing.SetAllowedVehiclesForIndex(allowed_vehicles, manager.NodeToIndex(idx))

    # Search parameters
    search_parameters = pywrapcp.DefaultRoutingSearchParameters()
    search_parameters.first_solution_strategy = (
        routing_enums_pb2.FirstSolutionStrategy.PATH_CHEAPEST_ARC)
    search_parameters.time_limit.seconds = 10

    solution = routing.SolveWithParameters(search_parameters)
    if not solution:
        print('No solution found!')
        return
    # Print and visualize solution
    routes = []
    for vehicle_id, vconf in enumerate(vehicle_configs):
        index = routing.Start(vehicle_id)
        plan = []
        route_load = 0
        while not routing.IsEnd(index):
            node = manager.IndexToNode(index)
            if node != 0:
                plan.append(node-1)  # -1 because orders are after depot
                route_load += demands[node]
            index = solution.Value(routing.NextVar(index))
        print(f'Route for vehicle {vehicle_id} ({vconf["type"]}): {[orders[i]["order_id"] for i in plan]} (Total load: {route_load}kg)')
        routes.append(plan)
    if visualize:
        plot_routes_with_depot_advanced(depot_city, orders, routes, CITY_COORDS, vehicle_configs)

def plot_routes_with_depot_advanced(depot_city, orders, routes, city_coords, vehicle_configs):
    plt.figure(figsize=(10, 7))
    depot_lat, depot_lon = city_coords[depot_city]
    plt.scatter([depot_lon], [depot_lat], marker='*', color='orange', s=250, label='Depot')
    for v, route in enumerate(routes):
        color = vehicle_configs[v].get("color", f'C{v}')
        xs, ys = [depot_lon], [depot_lat]
        for idx in route:
            city = orders[idx]['pickup_area']
            if city in city_coords:
                lat, lon = city_coords[city]
                xs.append(lon)
                ys.append(lat)
        plt.plot(xs, ys, marker='o', color=color, label=f'Vehicle {v} ({vehicle_configs[v]["type"]})')
        # Annotate orders with vehicle type and order id
        for idx in route:
            city = orders[idx]['pickup_area']
            if city in city_coords:
                lat, lon = city_coords[city]
                plt.annotate(f"{orders[idx]['order_id']}\n{orders[idx].get('required_vehicle_type','')}", (lon, lat), fontsize=8, color=color)
    plt.xlabel('Longitude')
    plt.ylabel('Latitude')
    plt.title(f'Optimized Delivery Routes (Depot: {depot_city})')
    plt.legend()
    plt.tight_layout()
    plt.show()

if __name__ == "__main__":
    print("--- OR-Tools CVRPTW Demo with Depot ---")
    depot = 'Port Blair'  # Change as needed
    random_orders = [random_order(i+1) for i in range(8)]
    # Example: 2 vehicles, one bike, one van/truck
    vehicle_configs = [
        {"type": "bike", "capacity": 15, "allowed_zones": None, "color": "b"},
        {"type": "van/truck", "capacity": 30, "allowed_zones": None, "color": "g"},
    ]
    solve_vrp_with_depot(random_orders, depot_city=depot, vehicle_configs=vehicle_configs, visualize=True)
