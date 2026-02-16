import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Navigation } from './components/Navigation';
import { Dashboard } from './components/Dashboard';
import { NormalMode } from './pages/NormalMode';
import { DrainMode } from './pages/DrainMode';
import { FreezeMode } from './pages/FreezeMode';
import { RecoveryMode } from './pages/RecoveryMode';
import { TenantProvider } from './contexts/TenantContext';

function App() {
  return (
    <TenantProvider>
      <Router>
        <div className="app-container">
          <Navigation />
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/normal" element={<NormalMode onSetMode={() => { }} />} />
            <Route path="/drain" element={<DrainMode />} />
            <Route path="/freeze" element={<FreezeMode />} />
            <Route path="/recovery" element={<RecoveryMode />} />
          </Routes>
        </div>
      </Router>
    </TenantProvider>
  );
}

export default App;
