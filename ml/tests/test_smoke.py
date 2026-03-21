from pathlib import Path


def test_ml_service_layout_exists() -> None:
    root = Path(__file__).resolve().parents[1]

    required_files = [
        root / "requirements.txt",
        root / "premium" / "main.py",
        root / "fraud" / "main.py",
        root / "forecast" / "main.py",
    ]

    for file_path in required_files:
        assert file_path.exists(), f"Missing required file: {file_path}"


def test_synthetic_data_present() -> None:
    root = Path(__file__).resolve().parents[1]

    required_datasets = [
        root / "premium" / "data" / "synthetic_training_data.csv",
        root / "fraud" / "data" / "synthetic_claim_patterns.csv",
        root / "forecast" / "data" / "synthetic_zone_history.csv",
    ]

    for dataset_path in required_datasets:
        assert dataset_path.exists(), f"Missing dataset: {dataset_path}"
        assert dataset_path.stat().st_size > 0, f"Dataset is empty: {dataset_path}"
