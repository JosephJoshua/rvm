/* @refresh reload */
import { render } from 'solid-js/web';
import { Router, Routes, Route } from '@solidjs/router';

import './index.css';

import ProtectedRoutes from './components/ProtectedRoutes';
import SignIn from './pages/SignIn';
import SignUp from './pages/SignUp';
import Home from './pages/Home';
import GuestOnlyRoutes from './components/GuestOnlyRoutes';
import Scan from './pages/Scan';

const root = document.getElementById('root');

render(
  () => (
    <Router>
      <Routes>
        <Route path="/" component={GuestOnlyRoutes}>
          <Route path="signin" component={SignIn} />
          <Route path="signup" component={SignUp} />
        </Route>

        <Route path="/" component={ProtectedRoutes}>
          <Route path="" component={Home} />
          <Route path="scan" component={Scan} />
        </Route>
      </Routes>
    </Router>
  ),
  root!,
);
