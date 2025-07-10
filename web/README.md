# GZH Monitoring Dashboard - React Frontend

A modern React-based single-page application (SPA) for the GZH Manager monitoring system.

## Features

- **Modern React SPA** - Built with React 18 and Material-UI
- **Real-time Updates** - WebSocket integration for live system monitoring
- **Authentication** - JWT-based authentication with role-based access control
- **Responsive Design** - Works seamlessly on desktop and mobile devices
- **Interactive Charts** - Real-time system metrics visualization
- **Task Management** - Monitor and control running tasks
- **Alert System** - Real-time alerts with severity levels
- **WebSocket Logs** - Live activity monitoring

## Tech Stack

- **React 18** - Modern React with hooks
- **Material-UI (MUI)** - Google's Material Design components
- **Recharts** - Charts and data visualization
- **Axios** - HTTP client for API communication
- **React Router** - Client-side routing
- **WebSocket** - Real-time communication

## Development Setup

### Prerequisites

- Node.js 16+ and npm
- Go backend server running on port 8080

### Installation

```bash
# Navigate to web directory
cd web

# Install dependencies
npm install

# Start development server
npm start
```

The development server will start on `http://localhost:3000` and proxy API requests to the Go backend on `http://localhost:8080`.

### Available Scripts

- `npm start` - Start development server with hot reload
- `npm run build` - Build for production
- `npm test` - Run test suite
- `npm run eject` - Eject from Create React App (⚠️ irreversible)

## Production Build

```bash
# Build the React app
npm run build

# The build artifacts will be in the 'build' directory
# The Go server will serve these files from ./web/build/
```

## Project Structure

```
web/
├── public/           # Static assets
├── src/
│   ├── components/   # React components
│   │   ├── Dashboard.js    # Main dashboard
│   │   ├── Layout.js       # App layout with navigation
│   │   ├── Login.js        # Authentication form
│   │   └── LoadingSpinner.js
│   ├── contexts/     # React contexts
│   │   ├── AuthContext.js      # Authentication state
│   │   └── WebSocketContext.js # WebSocket connection
│   ├── services/     # API and service layer
│   │   └── api.js           # HTTP API client
│   ├── App.js        # Main app component
│   └── index.js      # App entry point
├── package.json      # Dependencies and scripts
└── README.md        # This file
```

## Features Detail

### Authentication

The app implements JWT-based authentication with the following roles:

- **Admin** - Full access to all features including user management
- **Operator** - Can view and manage tasks, alerts, and configurations
- **Viewer** - Read-only access to dashboards and metrics

Default credentials:
- Admin: `admin` / `admin123`
- Viewer: `viewer` / `viewer123`

### Dashboard Components

1. **System Status Cards** - CPU, Memory, Tasks, and overall health
2. **Real-time Charts** - System metrics and network I/O visualization
3. **Task Manager** - View and control running tasks with progress tracking
4. **Alert Center** - Real-time alerts with severity-based filtering
5. **WebSocket Activity Log** - Live connection status and event monitoring

### WebSocket Integration

The frontend maintains a persistent WebSocket connection for:

- Real-time system status updates
- Task progress notifications
- Alert notifications
- System metric streaming
- Connection health monitoring with auto-reconnect

### Responsive Design

The dashboard is fully responsive and provides:

- Mobile-friendly navigation drawer
- Adaptive chart sizing
- Touch-friendly interface elements
- Progressive web app (PWA) support

## API Integration

The frontend communicates with the Go backend through:

- **REST API** - For authentication, configuration, and data retrieval
- **WebSocket** - For real-time updates and live monitoring
- **Authentication** - JWT tokens with automatic refresh handling

## Development Guidelines

### Adding New Components

1. Create component in `src/components/`
2. Add routing in `App.js` if needed
3. Update navigation in `Layout.js`
4. Implement proper error handling and loading states

### Adding New API Endpoints

1. Add method to `src/services/api.js`
2. Update authentication context if needed
3. Add proper error handling and validation

### Styling Guidelines

- Use Material-UI theme system for consistent styling
- Follow Material Design principles
- Implement responsive breakpoints
- Use proper color schemes for accessibility

## Deployment

The React app is built into static files that are served by the Go backend:

```bash
# Build the frontend
npm run build

# Start the Go backend (serves both API and frontend)
cd ..
make build
./gz monitoring server
```

The Go server will serve:
- React SPA from `/` (with routing fallback)
- API endpoints from `/api/v1/*`
- WebSocket from `/ws`
- Legacy dashboard from `/dashboard`

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Contributing

1. Follow React best practices and hooks patterns
2. Use TypeScript for type safety (future enhancement)
3. Write tests for critical components
4. Follow Material-UI design system
5. Maintain accessibility standards