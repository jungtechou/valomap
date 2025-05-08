# VALORANT Map Picker Frontend

A modern React.js application for randomly selecting VALORANT maps. This frontend connects to the backend API to provide a sleek, user-friendly interface for the map roulette functionality.

## Features

- Random map selection with beautiful animations
- Toggle to filter only standard maps (maps with tactical descriptions)
- Responsive design for all device sizes
- Elegant VALORANT-inspired UI
- Accessibility features
- Performance optimizations including memoization

## Technologies Used

- React.js 18
- Styled Components for CSS-in-JS styling
- Framer Motion for smooth animations
- Axios for API requests
- React Icons
- ESLint and Prettier for code quality
- Husky and lint-staged for pre-commit hooks

## Development

### Prerequisites

- Node.js 18 or later
- npm 8 or later

### Setup

1. Clone the repository:
```bash
git clone https://github.com/jungtechou/valomap.git
cd valomap
```

2. Install dependencies:
```bash
cd frontend
npm install
```

3. Start the development server:
```bash
npm start
```

The app will be available at http://localhost:3000 and will proxy API requests to the backend.

### Code Quality Tools

This project includes several tools to maintain code quality:

- **ESLint**: Lints JavaScript code
  ```bash
  npm run lint
  ```

- **Prettier**: Formats code consistently
  ```bash
  npm run format
  ```

- **Husky**: Runs linting and formatting before commits
  ```bash
  # Automatically runs on git commit
  ```

### Building for Production

```bash
npm run build
```

This will create an optimized production build in the `build` directory.

## Docker Deployment

The frontend includes a Dockerfile and nginx configuration for easy deployment:

```bash
# Build and run with Docker
docker build -t valomap-frontend .
docker run -p 80:80 valomap-frontend
```

When using Docker Compose, the frontend will automatically connect to the backend service as defined in the docker-compose.yml file.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b my-new-feature`
3. Make your changes and commit them: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request

Please ensure your code follows the project's coding standards by running ESLint and Prettier before submitting a PR.
