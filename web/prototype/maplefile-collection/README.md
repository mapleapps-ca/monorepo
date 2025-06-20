# MapleFile Collections - React Frontend

A modern React application for managing encrypted file collections using the MapleFile API. Built with clean architecture principles, dependency injection, and comprehensive CRUD operations.

## 🏗️ Architecture Overview

This application follows **Clean Architecture** principles with clear separation of concerns:

### Dependency Inversion Principle
- Services depend on abstractions, not concrete implementations
- Dependencies are injected through React Context
- Easy to test and maintain

### Single Responsibility Principle
- Each component, service, and hook has a single, well-defined purpose
- Clear separation between UI, business logic, and data access

### Directory Structure

```
src/
├── components/           # Reusable UI components
│   ├── ui/              # Basic UI components (Button, Input, Modal, etc.)
│   ├── collections/     # Collection-specific components
│   └── layout/          # Layout components (Navigation, Layout)
├── pages/               # Page components (one per route)
├── services/            # Business logic and API communication
├── hooks/               # Custom React hooks for state management
├── contexts/            # React Context for dependency injection
├── constants/           # Application constants and enums
├── utils/              # Utility functions
├── App.jsx             # Main application component with routing
└── main.jsx            # Application entry point
```

## 🚀 Features

### Core CRUD Operations
- ✅ **Create** collections (folders and albums)
- ✅ **Read** collections (list, detail, filter)
- ✅ **Update** collection metadata
- ✅ **Delete** collections (soft delete)

### Advanced Features
- ✅ **Share** collections with other users
- ✅ **Permissions** management (read-only, read-write, admin)
- ✅ **Hierarchical** collections (nested folders)
- ✅ **Encryption** support (client-side encryption simulation)
- ✅ **Real-time** state management
- ✅ **Error handling** with user-friendly messages
- ✅ **Loading states** and optimistic updates

### User Experience
- ✅ **Responsive design** (mobile-friendly)
- ✅ **Accessibility** features (keyboard navigation, screen readers)
- ✅ **Modern UI** with Tailwind CSS
- ✅ **Client-side routing** with React Router
- ✅ **Form validation** and error handling

## 🛠️ Technology Stack

### Core Technologies
- **React 19** - UI library with latest features
- **Vite** - Fast build tool and dev server
- **React Router 6** - Client-side routing
- **Tailwind CSS 3** - Utility-first CSS framework

### Development Tools
- **ESLint** - Code linting and formatting
- **PostCSS** - CSS processing
- **Autoprefixer** - CSS vendor prefixes

### Architecture Patterns
- **Dependency Injection** - Using React Context
- **Custom Hooks** - For state management and side effects
- **Service Layer** - For API communication and business logic
- **Component Composition** - Reusable and composable UI components

## 📦 Installation & Setup

### Prerequisites
- Node.js 18+ and npm/yarn
- Backend server running on `http://localhost:8000`

### Installation Steps

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd maplefile-collection
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Start the development server**
   ```bash
   npm run dev
   ```

4. **Open your browser**
   Navigate to `http://localhost:3000`

### Available Scripts

```bash
npm run dev      # Start development server
npm run build    # Build for production
npm run preview  # Preview production build
npm run lint     # Run ESLint
```

## 🔧 Configuration

### API Configuration
The application is configured to proxy API requests to `http://localhost:8000` in development. Update `vite.config.js` to change the backend URL:

```javascript
export default defineConfig({
  server: {
    proxy: {
      "/maplefile/api": {
        target: "http://your-backend-url",
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
```

### Environment Variables
Create a `.env` file for environment-specific configuration:

```env
VITE_API_BASE_URL=http://localhost:8000/maplefile/api/v1
VITE_APP_TITLE=MapleFile Collections
```

## 🏛️ Architecture Details

### Service Layer
The application uses a service-oriented architecture:

#### ApiService
- Base HTTP client with error handling
- Consistent request/response format
- Automatic JSON parsing and error handling

#### CollectionService
- All collection-related API operations
- Follows the MapleFile API specification
- Depends on ApiService abstraction

#### CryptoService
- Client-side encryption/decryption (simulated)
- Key generation and management
- In production, would use actual cryptographic libraries

### State Management
- **React Context** for dependency injection
- **Custom hooks** for component state
- **Service layer** for business logic
- No global state library needed due to clear architecture

### Error Handling
- Centralized error handling in services
- User-friendly error messages
- Graceful degradation for network issues
- Loading states for all async operations

### Security Considerations
- Client-side encryption simulation
- Input validation and sanitization
- Secure API communication patterns
- Permission-based UI rendering

## 🧪 Testing Strategy

### Component Testing
```bash
# Unit tests for components
npm run test:components

# Integration tests for pages
npm run test:integration
```

### Service Testing
```bash
# Unit tests for services
npm run test:services

# API integration tests
npm run test:api
```

### E2E Testing
```bash
# End-to-end tests
npm run test:e2e
```

## 📚 API Integration

The application integrates with the MapleFile Collections API with support for:

### Collection Operations
- Create, read, update, delete collections
- Hierarchical collection management
- Collection type support (folder/album)

### Sharing & Permissions
- Share collections with other users
- Fine-grained permission control
- Member management

### Synchronization
- Offline-first design considerations
- Sync API for data consistency
- Optimistic updates

## 🔍 Code Examples

### Creating a New Collection
```javascript
const { createCollection } = useCollectionOperations();

const handleSubmit = async (formData) => {
  try {
    const collection = await createCollection({
      name: formData.name,
      type: formData.type,
      parent_id: formData.parent_id,
    });
    navigate(`/collections/${collection.id}`);
  } catch (error) {
    // Error handling is automatic
  }
};
```

### Using the Collections Hook
```javascript
const { collections, loading, error, refreshCollections } = useCollections({
  include_owned: true,
  include_shared: false,
});

// Auto-refreshes when filters change
// Provides loading states and error handling
```

### Service Dependency Injection
```javascript
const { collectionService, cryptoService } = useServices();

// Services are injected and ready to use
// Easy to mock for testing
```

## 🚧 Future Enhancements

### Planned Features
- [ ] File upload and management
- [ ] Advanced search and filtering
- [ ] Bulk operations
- [ ] Collection templates
- [ ] Activity history and audit logs
- [ ] Real-time collaboration
- [ ] Offline support with service workers

### Technical Improvements
- [ ] Add comprehensive unit tests
- [ ] Implement E2E testing with Playwright
- [ ] Add Storybook for component documentation
- [ ] Performance optimization with React.memo
- [ ] Add analytics and monitoring
- [ ] Implement progressive web app features

## 🤝 Contributing

### Development Guidelines
1. Follow the existing architecture patterns
2. Maintain Single Responsibility Principle
3. Use TypeScript for new features (migration planned)
4. Write tests for new functionality
5. Follow the established code style

### Code Style
- Use functional components with hooks
- Prefer composition over inheritance
- Keep components small and focused
- Use descriptive variable and function names
- Add JSDoc comments for complex functions

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

### Common Issues
1. **CORS errors**: Ensure backend server is running and CORS is configured
2. **Build errors**: Clear node_modules and reinstall dependencies
3. **Styling issues**: Verify Tailwind CSS is properly configured

### Getting Help
- Check the console for error messages
- Review the API documentation
- Check network requests in browser dev tools
- Verify backend server is responding correctly

---

Built with ❤️ using React and modern web technologies.
