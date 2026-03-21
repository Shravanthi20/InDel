# Rule-based fraud detection

class RuleBasedDetector:
    def __init__(self):
        self.rules = []
    
    def detect(self, claim_data):
        violations = []
        
        # Rule 1: Duplicate claim within 7 days
        if claim_data.get('duplicate_claim_7d'):
            violations.append("duplicate_claim_recent")
        
        # Rule 2: Claim amount exceeds 3x baseline
        if claim_data.get('claim_amount', 0) > (claim_data.get('baseline', 0) * 3):
            violations.append("claim_exceeds_baseline")
        
        # Rule 3: Multiple claims same disruption
        if claim_data.get('claims_same_disruption', 0) > 1:
            violations.append("multiple_claims_disruption")
        
        return violations
