FROM python:3.9-slim

# Create app directory
WORKDIR /app

# Copy requirements and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Create data directory
RUN mkdir -p /app/data

# Expose Dash port
EXPOSE 8050

# Start the Dash application
CMD ["streamlit", "run", "dashboard.py", "--server.port=8050"]
