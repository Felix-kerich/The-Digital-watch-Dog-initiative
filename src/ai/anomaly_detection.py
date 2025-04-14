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
    'geo_distance'
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
            'geo_distance': np.random.exponential(scale=100, size=n_samples)
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
    else:
        previous_transaction_count = 0
        average_amount = transaction['amount']
        frequency = 0
        time_since_last = 0
    
    # Source and destination risk scores (placeholder - would come from a real risk scoring system)
    # In a real system, these would be based on historical behavior, KYC data, etc.
    source_risk_score = 0.1  # Low risk by default
    destination_risk_score = 0.1  # Low risk by default
    
    # Geographic distance (placeholder - would come from real geo data)
    geo_distance = 0
    
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
        'geo_distance': geo_distance
    }
    
    return features

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
    # Find the most anomalous features
    feature_list = list(features.columns)
    feature_values = features.iloc[0].values
    
    # Create a basic reason based on transaction amount
    if feature_values[feature_list.index('amount')] > 10000:
        return "Unusually large transaction amount"
    
    if feature_values[feature_list.index('time_since_last')] < 1 and feature_values[feature_list.index('previous_transaction_count')] > 0:
        return "Unusually frequent activity compared to historical pattern"
    
    if feature_values[feature_list.index('transaction_hour')] >= 22 or feature_values[feature_list.index('transaction_hour')] <= 5:
        return "Unusual transaction time (overnight hours)"
        
    # Generic reason
    return "Transaction patterns differ from normal behavior"

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