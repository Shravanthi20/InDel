"""
Enhanced Route Optimizer with Constraints and Visualization
- Supports time windows, vehicle types, batching
- Visualizes routes using matplotlib
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
from fake_order_publisher import random_order, CITY_STATE_LOOKUP

# --- Haversine distance ---
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

# --- Build distance matrix ---
def build_distance_matrix(orders: List[dict]) -> List[List[int]]:
    n = len(orders)
    matrix = [[0]*n for _ in range(n)]
    for i in range(n):
        for j in range(n):
            if i == j:
                matrix[i][j] = 0
            else:
                src = orders[i]['pickup_area']
                dst = orders[j]['drop_area']
                if src in CITY_COORDS and dst in CITY_COORDS:
                    matrix[i][j] = int(haversine(*CITY_COORDS[src], *CITY_COORDS[dst]) * 1000)
                else:
                    matrix[i][j] = random.randint(1000, 10000)
    return matrix

# --- Visualize routes ---
def plot_routes(orders, routes, city_coords):
    plt.figure(figsize=(8, 6))
    colors = ['b', 'g', 'r', 'c', 'm', 'y', 'k']
    for v, route in enumerate(routes):
        color = colors[v % len(colors)]
        xs, ys = [], []
        for idx in route:
            city = orders[idx]['pickup_area']
            if city in city_coords:
                lat, lon = city_coords[city]
                xs.append(lon)
                ys.append(lat)
        plt.plot(xs, ys, marker='o', color=color, label=f'Vehicle {v}')
    plt.xlabel('Longitude')
    plt.ylabel('Latitude')
    plt.title('Optimized Delivery Routes')
    plt.legend()
    plt.show()

# --- Enhanced OR-Tools CVRPTW Solver ---
def solve_vrp(orders: List[dict], num_vehicles: int = 2, depot: int = 0, visualize: bool = True):
    dist_matrix = build_distance_matrix(orders)
    manager = pywrapcp.RoutingIndexManager(len(dist_matrix), num_vehicles, depot)
    routing = pywrapcp.RoutingModel(manager)

    def distance_callback(from_index, to_index):
        from_node = manager.IndexToNode(from_index)
        to_node = manager.IndexToNode(to_index)
        return dist_matrix[from_node][to_node]

    transit_callback_index = routing.RegisterTransitCallback(distance_callback)
    routing.SetArcCostEvaluatorOfAllVehicles(transit_callback_index)

    # Capacity constraint
    demands = [int(order['package_weight_kg']) for order in orders]
    vehicle_capacities = [30] * num_vehicles
    def demand_callback(from_index):
        from_node = manager.IndexToNode(from_index)
        return demands[from_node]
    demand_callback_index = routing.RegisterUnaryTransitCallback(demand_callback)
    routing.AddDimensionWithVehicleCapacity(
        demand_callback_index, 0, vehicle_capacities, True, 'Capacity')

    # Time window constraint (random for demo)
    time_windows = [(0, 10000)] * len(orders)
    for i, order in enumerate(orders):
        # For demo, randomize time windows
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

    # Vehicle type constraint (van/truck for inter-state)
    vehicle_types = ['bike/small van', 'van/truck']
    order_vehicle_types = [order.get('required_vehicle_type', 'bike/small van') for order in orders]
    # Only allow vehicle 0 to take bike/small van, vehicle 1 can take any
    for idx, req_type in enumerate(order_vehicle_types):
        if req_type == 'van/truck':
            routing.SetAllowedVehiclesForIndex([1], manager.NodeToIndex(idx))

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
    for vehicle_id in range(num_vehicles):
        index = routing.Start(vehicle_id)
        plan = []
        route_load = 0
        while not routing.IsEnd(index):
            node = manager.IndexToNode(index)
            plan.append(node)
            route_load += demands[node]
            index = solution.Value(routing.NextVar(index))
        print(f'Route for vehicle {vehicle_id}: {[orders[i]["order_id"] for i in plan]} (Total load: {route_load}kg)')
        routes.append(plan)
    if visualize:
        plot_routes(orders, routes, CITY_COORDS)

if __name__ == "__main__":
    print("--- Enhanced OR-Tools CVRPTW Demo: Random Orders ---")
    random_orders = [random_order(i+1) for i in range(8)]
    solve_vrp(random_orders, num_vehicles=2, visualize=True)

    print("\n--- Enhanced OR-Tools CVRPTW Demo: Fixed Sample Batch ---")
    sample_orders = [
        {'order_id': 'A', 'pickup_area': 'Port Blair', 'drop_area': 'Addanki', 'package_weight_kg': 5, 'required_vehicle_type': 'van/truck'},
        {'order_id': 'B', 'pickup_area': 'Addanki', 'drop_area': 'Port Blair', 'package_weight_kg': 7, 'required_vehicle_type': 'van/truck'},
        {'order_id': 'C', 'pickup_area': 'Amalapuram', 'drop_area': 'Anantapur', 'package_weight_kg': 3, 'required_vehicle_type': 'bike/small van'},
        {'order_id': 'D', 'pickup_area': 'Anantapur', 'drop_area': 'Amalapuram', 'package_weight_kg': 4, 'required_vehicle_type': 'bike/small van'},
    ]
    solve_vrp(sample_orders, num_vehicles=2, visualize=True)
