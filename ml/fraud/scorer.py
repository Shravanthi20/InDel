from isolation_forest import IsolationForestDetector
from dbscan import DBSCANDetector
from rules import RuleBasedDetector

class FraudScorer:
    def __init__(self):
        self.iso_forest = IsolationForestDetector()
        self.dbscan = DBSCANDetector()
        self.rules = RuleBasedDetector()
    
    def score(self, claim_data):
        # Layer 1: Isolation Forest
        iso_score = self.iso_forest.detect(claim_data)
        
        # Layer 2: DBSCAN
        cluster = self.dbscan.detect(claim_data)
        
        # Layer 3: Rules
        violations = self.rules.detect(claim_data)
        
        # Combine scores
        final_score = (iso_score * 0.4) + (len(violations) * 0.2)
        
        verdict = "clean" if final_score < 0.5 else "suspicious"
        
        return {
            "fraud_score": final_score,
            "verdict": verdict,
            "signals": violations
        }
