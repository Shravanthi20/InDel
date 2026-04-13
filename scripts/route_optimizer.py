"""
Route Optimizer using OR-Tools CVRPTW for Delivery Orders
- Can use random orders (from fake_order_publisher) or a fixed sample batch
- Uses straight-line (Haversine) distance between cities as cost
"""
import sys
import os
import random
from typing import List, Dict
from ortools.constraint_solver import routing_enums_pb2
from ortools.constraint_solver import pywrapcp
import math

# Import order generator from fake_order_publisher
sys.path.append(os.path.dirname(__file__))
from fake_order_publisher import random_order, CITY_STATE_LOOKUP

# --- Helper: Haversine distance (km) ---
def haversine(lat1, lon1, lat2, lon2):
    R = 6371  # Earth radius in km
    phi1, phi2 = math.radians(lat1), math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lon2 - lon1)
    a = math.sin(dphi/2)**2 + math.cos(phi1)*math.cos(phi2)*math.sin(dlambda/2)**2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))

# --- Load city coordinates from CSV ---
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

# --- Build distance matrix for orders ---
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
                    matrix[i][j] = int(haversine(*CITY_COORDS[src], *CITY_COORDS[dst]) * 1000)  # meters
                else:
                    matrix[i][j] = random.randint(1000, 10000)  # fallback
    return matrix

# --- OR-Tools CVRPTW Solver ---
def solve_vrp(orders: List[dict], num_vehicles: int = 2, depot: int = 0):
    dist_matrix = build_distance_matrix(orders)
    manager = pywrapcp.RoutingIndexManager(len(dist_matrix), num_vehicles, depot)
    routing = pywrapcp.RoutingModel(manager)

    def distance_callback(from_index, to_index):
        from_node = manager.IndexToNode(from_index)
        to_node = manager.IndexToNode(to_index)
        return dist_matrix[from_node][to_node]

    transit_callback_index = routing.RegisterTransitCallback(distance_callback)
    routing.SetArcCostEvaluatorOfAllVehicles(transit_callback_index)

    # Add capacity constraint (use package_weight_kg)
    demands = [int(order['package_weight_kg']) for order in orders]
    vehicle_capacities = [30] * num_vehicles  # Example: 30kg per vehicle
    def demand_callback(from_index):
        from_node = manager.IndexToNode(from_index)
        return demands[from_node]
    demand_callback_index = routing.RegisterUnaryTransitCallback(demand_callback)
    routing.AddDimensionWithVehicleCapacity(
        demand_callback_index, 0, vehicle_capacities, True, 'Capacity')

    # Add time window constraint (dummy for now)
    time_windows = [(0, 10000)] * len(orders)
    def time_callback(from_index, to_index):
        return dist_matrix[manager.IndexToNode(from_index)][manager.IndexToNode(to_index)] // 10  # 1/10th of distance as time
    time_callback_index = routing.RegisterTransitCallback(time_callback)
    routing.AddDimension(
        time_callback_index, 1000, 100000, False, 'Time')
    time_dimension = routing.GetDimensionOrDie('Time')
    for idx, window in enumerate(time_windows):
        index = manager.NodeToIndex(idx)
        time_dimension.CumulVar(index).SetRange(*window)

    # Search parameters
    search_parameters = pywrapcp.DefaultRoutingSearchParameters()
    search_parameters.first_solution_strategy = (
        routing_enums_pb2.FirstSolutionStrategy.PATH_CHEAPEST_ARC)
    search_parameters.time_limit.seconds = 10

    # Solve
    solution = routing.SolveWithParameters(search_parameters)
    if not solution:
        print('No solution found!')
        return
    # Print solution
    for vehicle_id in range(num_vehicles):
        index = routing.Start(vehicle_id)
        plan = []
        route_load = 0
        while not routing.IsEnd(index):
            node = manager.IndexToNode(index)
            plan.append(orders[node]['order_id'])
            route_load += demands[node]
            index = solution.Value(routing.NextVar(index))
        print(f'Route for vehicle {vehicle_id}: {plan} (Total load: {route_load}kg)')

if __name__ == "__main__":
    print("--- OR-Tools CVRPTW Demo: Random Orders ---")
    random_orders = [random_order(i+1) for i in range(8)]
    solve_vrp(random_orders, num_vehicles=2)

    print("\n--- OR-Tools CVRPTW Demo: Fixed Sample Batch ---")
    sample_orders = [
        {'order_id': 'A', 'pickup_area': 'Port Blair', 'drop_area': 'Addanki', 'package_weight_kg': 5},
        {'order_id': 'B', 'pickup_area': 'Addanki', 'drop_area': 'Port Blair', 'package_weight_kg': 7},
        {'order_id': 'C', 'pickup_area': 'Amalapuram', 'drop_area': 'Anantapur', 'package_weight_kg': 3},
        {'order_id': 'D', 'pickup_area': 'Anantapur', 'drop_area': 'Amalapuram', 'package_weight_kg': 4},
    ]
    solve_vrp(sample_orders, num_vehicles=2)
