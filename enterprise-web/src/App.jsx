import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthProvider';
import ErrorBoundary from './components/ErrorBoundary/ErrorBoundary';
import Navbar from './components/Navbar/Navbar';
import ProtectedRoute from './components/ProtectedRoute/ProtectedRoute';
import HomePage from './pages/HomePage/HomePage';
import NotFoundPage from './pages/NotFoundPage/NotFoundPage';

// 路由级代码分割
const LoginPage = lazy(() => import('./pages/LoginPage/LoginPage'));
const RegisterPage = lazy(() => import('./pages/RegisterPage/RegisterPage'));
const TasksPage = lazy(() => import('./pages/tasks/TasksPage'));
const TaskDetailPage = lazy(() => import('./pages/tasks/TaskDetailPage'));
const ProfilePage = lazy(() => import('./pages/ProfilePage/ProfilePage'));

// 懒加载时的回退组件
function PageFallback() {
  return (
    <div style={{
      minHeight: 'calc(100vh - 60px)',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      background: '#f0f2f5',
      color: '#888',
      fontSize: 16,
    }}>
      加载中...
    </div>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <AuthProvider>
          <Navbar />
          <Suspense fallback={<PageFallback />}>
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="/tasks" element={
                <ProtectedRoute><TasksPage /></ProtectedRoute>
              } />
              <Route path="/tasks/:id" element={
                <ProtectedRoute><TaskDetailPage /></ProtectedRoute>
              } />
              <Route path="/profile" element={
                <ProtectedRoute><ProfilePage /></ProtectedRoute>
              } />
              {/* 404 兜底路由 */}
              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </Suspense>
        </AuthProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
