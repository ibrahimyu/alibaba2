#!/bin/bash

echo "Setting up Restaurant Video Generator..."

# Create necessary directories
mkdir -p backend/output
mkdir -p frontend/src/components

# Copy .env if it doesn't exist
if [ ! -f backend/.env ]; then
  echo "Creating sample .env file in backend directory..."
  cat > backend/.env << EOF
# API Keys for Restaurant Video Generator
OPENAI_API_KEY=your_openai_api_key_here
ALIBABA_API_KEY=your_alibaba_api_key_here
EOF
  echo "Please edit backend/.env with your actual API keys"
fi

# Install backend dependencies
echo "Installing backend dependencies..."
cd backend
npm install

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd ../frontend
npm install

echo "Setup complete!"
echo ""
echo "To start the application:"
echo "1. Make sure your API keys are set in backend/.env"
echo "2. Run 'cd backend && npm run dev' to start both backend and frontend"
echo ""
echo "The frontend will be available at: http://localhost:3001"
