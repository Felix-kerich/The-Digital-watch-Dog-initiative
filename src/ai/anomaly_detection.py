import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
import joblib
import os
from datetime import datetime
import json
from flask import Flask, request, jsonify
import pickle

app = Flask(__name__)

# Model file paths
MODEL_DIR = os.path.join(os.path.dirname(__file__), 'models')
ISOLATION_FOREST_MODEL = os.path.join(MODEL_DIR, 'isolation_forest_model.pkl')
SCALER_MODEL = os.path.join(MODEL_DIR, 'scaler_model.pkl')

# Ensure models directory exists
os.makedirs(MODEL_DIR, exist_ok=True)

# Features used for anomaly detection
FEATURES = [
    'amount', 
    'previous_transaction_count',
    'average_amount', 
    'frequency', 
    'time_since_last',
    'transaction_hour',
    'weekend',
    'source_risk_score',
    'destination_risk_score',
    'geo_distance',
    'category_avg_amount',
    'category_max_amount',
    'category_frequency',
    'utilization_rate',
    'fiscal_progress'
]

def load_or_train_model():
    """Load existing model or train a new one if it doesn't exist."""
    try:
        # Try to load existing model
        model = joblib.load(ISOLATION_FOREST_MODEL)
        scaler = joblib.load(SCALER_MODEL)
        print("Loaded existing anomaly detection model")
        return model, scaler
    except:
        print("Training new anomaly detection model")
        # Create dummy data for initial training (will be replaced with real data in production)
        np.random.seed(42)
        n_samples = 1000
        
        # Generate synthetic transaction data
        data = {
            'amount': np.random.exponential(scale=5000, size=n_samples),
            'previous_transaction_count': np.random.poisson(lam=5, size=n_samples),
            'average_amount': np.random.exponential(scale=5000, size=n_samples),
            'frequency': np.random.exponential(scale=7, size=n_samples),
            'time_since_last': np.random.exponential(scale=72, size=n_samples),
            'transaction_hour': np.random.randint(0, 24, size=n_samples),
            'weekend': np.random.choice([0, 1], size=n_samples, p=[0.7, 0.3]),
            'source_risk_score': np.random.uniform(0, 1, size=n_samples),
            'destination_risk_score': np.random.uniform(0, 1, size=n_samples),
            'geo_distance': np.random.exponential(scale=100, size=n_samples),
            'category_avg_amount': np.random.exponential(scale=5000, size=n_samples),
            'category_max_amount': np.random.exponential(scale=10000, size=n_samples),
            'category_frequency': np.random.exponential(scale=5, size=n_samples),
            'utilization_rate': np.random.uniform(0, 1, size=n_samples),
            'fiscal_progress': np.random.uniform(0, 1, size=n_samples)
        }
        
        # Create DataFrame
        df = pd.DataFrame(data)
        
        # Standardize features
        scaler = StandardScaler()
        df_scaled = scaler.fit_transform(df[FEATURES])
        
        # Train Isolation Forest model
        model = IsolationForest(
            n_estimators=100,
            max_samples='auto',
            contamination=0.05,  # Assuming 5% of transactions are anomalous
            random_state=42
        )
        model.fit(df_scaled)
        
        # Save models
        joblib.dump(model, ISOLATION_FOREST_MODEL)
        joblib.dump(scaler, SCALER_MODEL)
        
        return model, scaler

# Load or train the model at startup
model, scaler = load_or_train_model()

def extract_features(transaction, historical_data=None):
    """Extract features from transaction data for anomaly detection."""
    if historical_data is None:
        historical_data = []
    
    # Parse transaction datetime
    txn_time = datetime.fromisoformat(transaction['timestamp'].replace('Z', '+00:00'))
    
    # Calculate transaction hour and weekend flag
    transaction_hour = txn_time.hour
    weekend = 1 if txn_time.weekday() >= 5 else 0
    
    # Get fund details
    fund_category = transaction.get('fund_category', 'UNKNOWN')
    fiscal_year = transaction.get('fiscal_year', '')
    budget_allocated = float(transaction.get('budget_allocated', 0))
    budget_utilized = float(transaction.get('budget_utilized', 0))
    
    # Calculate budget utilization rate
    utilization_rate = (budget_utilized / budget_allocated) if budget_allocated > 0 else 0
    
    # Calculate historical features
    if historical_data:
        df_hist = pd.DataFrame(historical_data)
        previous_transaction_count = len(df_hist)
        average_amount = df_hist['amount'].mean() if 'amount' in df_hist else transaction['amount']
        
        # Calculate frequency (transactions per week) based on historical data
        if len(df_hist) > 1 and 'timestamp' in df_hist:
            df_hist['timestamp'] = pd.to_datetime(df_hist['timestamp'])
            time_range_days = (df_hist['timestamp'].max() - df_hist['timestamp'].min()).total_seconds() / (24*3600)
            frequency = len(df_hist) / (time_range_days/7) if time_range_days > 0 else 0
        else:
            frequency = 0
        
        # Time since last transaction (in hours)
        if 'timestamp' in df_hist:
            last_txn_time = pd.to_datetime(df_hist['timestamp'].max())
            time_since_last = (txn_time - last_txn_time).total_seconds() / 3600
        else:
            time_since_last = 0
            
        # Calculate category spending patterns
        if 'fund_category' in df_hist:
            category_txns = df_hist[df_hist['fund_category'] == fund_category]
            category_avg = category_txns['amount'].mean() if len(category_txns) > 0 else transaction['amount']
            category_max = category_txns['amount'].max() if len(category_txns) > 0 else transaction['amount']
            category_frequency = len(category_txns) / (time_range_days/7) if time_range_days > 0 else 0
        else:
            category_avg = transaction['amount']
            category_max = transaction['amount']
            category_frequency = 0
    else:
        previous_transaction_count = 0
        average_amount = transaction['amount']
        frequency = 0
        time_since_last = 0
        category_avg = transaction['amount']
        category_max = transaction['amount']
        category_frequency = 0
    
    # Calculate fiscal year progress (0-1)
    try:
        fiscal_start = datetime.strptime(f"{fiscal_year}-07-01", "%Y-%m-%d")  # Assuming July-June fiscal year
        fiscal_end = fiscal_start.replace(year=fiscal_start.year + 1)
        fiscal_progress = (txn_time - fiscal_start).total_seconds() / (fiscal_end - fiscal_start).total_seconds()
    except:
        fiscal_progress = 0.5  # Default to mid-year if fiscal year parsing fails
    
    # Source and destination risk scores
    source_risk_score = calculate_entity_risk_score(transaction.get('source_id', ''), historical_data)
    destination_risk_score = calculate_entity_risk_score(transaction.get('destination_id', ''), historical_data)
    
    # Create feature vector
    features = {
        'amount': transaction['amount'],
        'previous_transaction_count': previous_transaction_count,
        'average_amount': average_amount,
        'frequency': frequency,
        'time_since_last': time_since_last,
        'transaction_hour': transaction_hour,
        'weekend': weekend,
        'source_risk_score': source_risk_score,
        'destination_risk_score': destination_risk_score,
        'category_avg_amount': category_avg,
        'category_max_amount': category_max,
        'category_frequency': category_frequency,
        'utilization_rate': utilization_rate,
        'fiscal_progress': fiscal_progress
    }
    
    return features

def calculate_entity_risk_score(entity_id, historical_data):
    """Calculate risk score for an entity based on historical behavior."""
    if not historical_data or not entity_id:
        return 0.1  # Default low risk
        
    df_hist = pd.DataFrame(historical_data)
    entity_txns = df_hist[
        (df_hist['source_id'] == entity_id) | 
        (df_hist['destination_id'] == entity_id)
    ]
    
    if len(entity_txns) == 0:
        return 0.2  # Slightly higher risk for new entities
    
    # Calculate risk factors
    avg_amount = entity_txns['amount'].mean()
    max_amount = entity_txns['amount'].max()
    txn_frequency = len(entity_txns) / ((entity_txns['timestamp'].max() - entity_txns['timestamp'].min()).total_seconds() / (24*3600*7))
    amount_volatility = entity_txns['amount'].std() / avg_amount if avg_amount > 0 else 0
    
    # Combine risk factors (simplified scoring)
    risk_score = (
        0.3 * min(1.0, max_amount / 1000000) +  # High amounts increase risk
        0.2 * min(1.0, txn_frequency / 10) +    # High frequency increases risk
        0.5 * min(1.0, amount_volatility)       # High volatility increases risk
    )
    
    return min(1.0, max(0.1, risk_score))

def detect_anomaly(transaction, historical_data=None):
    """Detect if a transaction is anomalous."""
    # Extract features
    features = extract_features(transaction, historical_data)
    
    # Convert to DataFrame and only select the features used by the model
    df = pd.DataFrame([features])
    df_features = df[FEATURES]
    
    # Scale features
    df_scaled = scaler.transform(df_features)
    
    # Predict anomaly
    prediction = model.predict(df_scaled)[0]
    anomaly_score = model.score_samples(df_scaled)[0]
    
    # Convert prediction: -1 is anomaly, 1 is normal
    is_anomaly = prediction == -1
    
    # Generate reason if it's an anomaly
    if is_anomaly:
        reason = generate_anomaly_reason(df_features, anomaly_score)
    else:
        reason = None
    
    return {
        'is_anomaly': is_anomaly,
        'anomaly_score': float(anomaly_score),
        'reason': reason
    }

def generate_anomaly_reason(features, score):
    """Generate a human-readable reason for why transaction was flagged."""
    reasons = []
    
    # Amount-based flags
    if features['amount'].iloc[0] > features['category_max_amount'].iloc[0] * 1.5:
        reasons.append("Transaction amount significantly higher than category maximum")
    
    if features['amount'].iloc[0] > features['average_amount'].iloc[0] * 3:
        reasons.append("Transaction amount unusually high compared to historical average")
    
    # Timing-based flags
    if features['transaction_hour'].iloc[0] >= 22 or features['transaction_hour'].iloc[0] <= 5:
        reasons.append("Unusual transaction time (overnight hours)")
    
    if features['weekend'].iloc[0] == 1:
        reasons.append("Weekend transaction unusual for government funds")
    
    # Budget utilization flags
    if features['utilization_rate'].iloc[0] > 0.9 and features['fiscal_progress'].iloc[0] < 0.5:
        reasons.append("High budget utilization rate early in fiscal year")
    
    if features['utilization_rate'].iloc[0] > 1.0:
        reasons.append("Transaction would exceed allocated budget")
    
    # Frequency-based flags
    if features['time_since_last'].iloc[0] < 24 and features['previous_transaction_count'].iloc[0] > 0:
        reasons.append("Unusually frequent transactions")
    
    # Risk score flags
    if features['source_risk_score'].iloc[0] > 0.7:
        reasons.append("Source entity has high risk score")
    
    if features['destination_risk_score'].iloc[0] > 0.7:
        reasons.append("Destination entity has high risk score")
    
    if not reasons:
        reasons.append("General anomalous pattern detected")
    
    return " | ".join(reasons)

@app.route('/detect', methods=['POST'])
def api_detect_anomaly():
    """API endpoint for anomaly detection."""
    if not request.json:
        return jsonify({'error': 'Request must be JSON'}), 400
    
    transaction = request.json.get('transaction')
    historical_data = request.json.get('historical_data', [])
    
    if not transaction:
        return jsonify({'error': 'Missing transaction data'}), 400
    
    required_fields = ['id', 'amount', 'timestamp']
    for field in required_fields:
        if field not in transaction:
            return jsonify({'error': f'Missing required field: {field}'}), 400
    
    result = detect_anomaly(transaction, historical_data)
    
    return jsonify({
        'transaction_id': transaction['id'],
        'is_anomaly': result['is_anomaly'],
        'anomaly_score': result['anomaly_score'],
        'reason': result['reason']
    })

@app.route('/retrain', methods=['POST'])
def api_retrain_model():
    """API endpoint to retrain the model with new data."""
    if not request.json:
        return jsonify({'error': 'Request must be JSON'}), 400
    
    transactions = request.json.get('transactions', [])
    
    if not transactions or len(transactions) < 100:
        return jsonify({'error': 'Need at least 100 transactions for retraining'}), 400
    
    try:
        # Extract features from each transaction
        features_list = []
        for txn in transactions:
            if all(field in txn for field in ['id', 'amount', 'timestamp']):
                features = extract_features(txn)
                features_list.append(features)
        
        # Convert to DataFrame
        df = pd.DataFrame(features_list)
        
        # Filter to only include required features
        df_features = df[FEATURES]
        
        # Standardize features
        scaler = StandardScaler()
        df_scaled = scaler.fit_transform(df_features)
        
        # Train new model
        model = IsolationForest(
            n_estimators=100,
            max_samples='auto',
            contamination=0.05,
            random_state=42
        )
        model.fit(df_scaled)
        
        # Save models
        joblib.dump(model, ISOLATION_FOREST_MODEL)
        joblib.dump(scaler, SCALER_MODEL)
        
        # Update global model
        globals()['model'] = model
        globals()['scaler'] = scaler
        
        return jsonify({'success': True, 'message': 'Model retrained successfully'})
    
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint."""
    return jsonify({
        'status': 'healthy',
        'model_loaded': model is not None and scaler is not None
    })

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 5000))
    app.run(host='0.0.0.0', port=port)