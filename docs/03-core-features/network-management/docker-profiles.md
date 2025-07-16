# Docker Network Profile Management

The Docker network profile management feature in `gz net-env docker-network` provides comprehensive container-specific network configuration capabilities. This feature allows you to manage Docker networks and container configurations as profiles, enabling easy deployment and management of complex containerized applications.

## Overview

Docker network profiles allow you to:
- Define and manage custom Docker networks with specific configurations
- Configure container-specific network settings including ports, DNS, and network aliases
- Import existing Docker Compose configurations
- Apply network profiles to create consistent development and production environments
- Manage container network isolation and segmentation

## Basic Usage

### Creating a Network Profile

```bash
# Create a simple network profile
gz net-env docker-network create myapp \
  --description "My application network profile" \
  --network app-network \
  --driver bridge \
  --subnet 172.20.0.0/16

# Create a profile interactively
gz net-env docker-network create myapp --interactive
```

### Managing Containers in Profiles

```bash
# Add a container to a profile
gz net-env docker-network container add myapp web \
  --image nginx:alpine \
  --network app-network \
  --port 80:80 \
  --port 443:443 \
  --env NGINX_HOST=example.com \
  --dns 8.8.8.8 \
  --alias web-server \
  --hostname web

# Update container configuration
gz net-env docker-network container update myapp web \
  --image nginx:latest \
  --port 8080:80

# Show container configuration
gz net-env docker-network container show myapp web

# Remove a container from profile
gz net-env docker-network container remove myapp web
```

### Applying Profiles

```bash
# Apply a network profile (creates networks and containers)
gz net-env docker-network apply myapp

# Dry run to see what would be applied
gz net-env docker-network apply myapp --dry-run
```

### Import from Docker Compose

```bash
# Import existing docker-compose.yml
gz net-env docker-network import ./docker-compose.yml myapp-imported
```

## Advanced Features

### Container Network Isolation

Create isolated network segments for security:

```yaml
name: secure-app
description: Application with network isolation
networks:
  public:
    driver: bridge
    subnet: 172.30.0.0/24
  internal:
    driver: bridge
    subnet: 172.31.0.0/24
    options:
      com.docker.network.bridge.enable_icc: "false"
  database:
    driver: bridge
    subnet: 172.32.0.0/24
    
containers:
  nginx:
    image: nginx:alpine
    networks:
      - public
    ports:
      - "80:80"
      - "443:443"
  
  api:
    image: myapp/api:latest
    networks:
      - public
      - internal
      - database
    environment:
      DB_HOST: postgres
      CACHE_HOST: redis
  
  postgres:
    image: postgres:13
    networks:
      - database
    environment:
      POSTGRES_PASSWORD: secret
  
  redis:
    image: redis:6-alpine
    networks:
      - database
```

### Container-Specific Network Settings

Each container can have detailed network configurations:

```yaml
containers:
  web:
    image: nginx:alpine
    network_mode: bridge  # or host, none, container:name
    networks:
      - frontend
      - backend
    ports:
      - "80:80"
      - "443:443"
    environment:
      NGINX_HOST: example.com
      NGINX_PORT: "80"
    dns:
      - 8.8.8.8
      - 8.8.4.4
    dns_search:
      - example.com
      - internal.example.com
    extra_hosts:
      - "api.local:172.20.0.10"
      - "db.local:172.20.0.20"
    hostname: web-server
    domainname: example.com
    network_alias:
      - web
      - nginx
      - frontend-server
```

### Profile Management Commands

```bash
# List all profiles
gz net-env docker-network list

# Export profile to file
gz net-env docker-network export myapp myapp-profile.yaml

# Clone an existing profile
gz net-env docker-network clone production staging

# Delete a profile
gz net-env docker-network delete myapp
```

### Network and Container Status

```bash
# Show Docker network status
gz net-env docker-network status

# Show networks and containers
gz net-env docker-network status --containers

# Get JSON output
gz net-env docker-network status --output json
```

### Runtime Container Network Management

```bash
# Connect a running container to a network
gz net-env docker-network container connect mycontainer mynetwork \
  --alias container-alias

# Disconnect a container from a network
gz net-env docker-network container disconnect mycontainer mynetwork
```

## Example Profiles

### Microservices Architecture

```yaml
name: microservices
description: Microservices application stack
networks:
  edge:
    driver: bridge
    subnet: 172.20.0.0/24
    labels:
      tier: edge
  
  services:
    driver: overlay
    subnet: 172.21.0.0/24
    attachable: true
    labels:
      tier: application
  
  data:
    driver: bridge
    subnet: 172.22.0.0/24
    options:
      com.docker.network.bridge.enable_icc: "false"
    labels:
      tier: data

containers:
  api-gateway:
    image: kong:latest
    networks:
      - edge
      - services
    ports:
      - "8000:8000"
      - "8443:8443"
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /kong/kong.yml
    network_alias:
      - gateway
      - api
  
  user-service:
    image: myapp/user-service:latest
    networks:
      - services
      - data
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis
    network_alias:
      - users
  
  order-service:
    image: myapp/order-service:latest
    networks:
      - services
      - data
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis
    network_alias:
      - orders
  
  postgres:
    image: postgres:13
    networks:
      - data
    environment:
      POSTGRES_DB: microservices
      POSTGRES_USER: app
      POSTGRES_PASSWORD: secret
    network_alias:
      - db
      - postgres
  
  redis:
    image: redis:6-alpine
    networks:
      - data
    network_alias:
      - cache
      - redis

compose:
  file: /path/to/docker-compose.yml
  project: microservices
  auto_apply: true
  environment:
    COMPOSE_PROJECT_NAME: microservices
    DOCKER_BUILDKIT: "1"
```

### Development Environment

```yaml
name: dev-env
description: Local development environment
networks:
  dev-net:
    driver: bridge
    subnet: 172.25.0.0/16

containers:
  mysql:
    image: mysql:8
    networks:
      - dev-net
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: devpass
      MYSQL_DATABASE: devdb
    network_alias:
      - db
      - mysql
  
  phpmyadmin:
    image: phpmyadmin:latest
    networks:
      - dev-net
    ports:
      - "8080:80"
    environment:
      PMA_HOST: mysql
      PMA_PORT: "3306"
  
  mailhog:
    image: mailhog/mailhog:latest
    networks:
      - dev-net
    ports:
      - "1025:1025"
      - "8025:8025"
    network_alias:
      - mail
      - smtp

metadata:
  environment: development
  team: backend
  version: "1.0"
```

## Best Practices

1. **Network Segmentation**: Use separate networks for different tiers (frontend, backend, database)
2. **Container Aliases**: Use meaningful network aliases for service discovery
3. **Environment Variables**: Store configuration in environment variables
4. **Port Management**: Document and standardize port mappings
5. **DNS Configuration**: Use custom DNS for internal service resolution
6. **Profile Versioning**: Include version metadata in profiles
7. **Security**: Disable inter-container communication (ICC) for sensitive networks

## Integration with Docker Compose

The system can import existing Docker Compose files and convert them to network profiles:

```bash
# Import docker-compose.yml
gz net-env docker-network import ./docker-compose.yml myapp

# The imported profile includes:
# - All networks defined in the compose file
# - All services as containers with their configurations
# - Port mappings, environment variables, and network connections
# - Reference to the original compose file for updates
```

## Troubleshooting

### Common Issues

1. **Network Already Exists**
   - The system checks for existing networks before creation
   - Use different network names or remove existing networks

2. **Container Connection Failed**
   - Ensure the network exists before connecting containers
   - Check that the container is running

3. **DNS Not Updated**
   - DNS settings cannot be changed on running containers
   - Recreate the container with new DNS settings

4. **Port Conflicts**
   - Check for existing port bindings before applying profiles
   - Use different host ports or stop conflicting containers

### Debugging

```bash
# Check profile details
gz net-env docker-network list -o yaml

# Verify network creation
docker network ls

# Inspect container network settings
docker inspect <container-name>

# Test network connectivity
docker exec <container> ping <other-container>
```

## Security Considerations

1. **Network Isolation**: Use separate networks for different security zones
2. **ICC (Inter-Container Communication)**: Disable ICC for sensitive networks
3. **Port Exposure**: Only expose necessary ports to the host
4. **Environment Variables**: Avoid storing secrets directly in profiles
5. **Network Policies**: Consider using Docker Swarm or Kubernetes for advanced policies

## Future Enhancements

- Support for Docker Swarm network configurations
- Integration with Kubernetes NetworkPolicies
- Network traffic analysis and reporting
- Automatic network topology visualization
- Integration with service mesh solutions (Istio, Linkerd)