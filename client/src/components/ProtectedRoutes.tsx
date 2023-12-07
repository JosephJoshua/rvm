import { Outlet, useNavigate } from '@solidjs/router';
import { onCleanup, onMount } from 'solid-js';
import { auth } from '../lib/firebase';
import { onAuthStateChanged } from 'firebase/auth';

const ProtectedRoutes = () => {
  const navigate = useNavigate();
  let unsubscribe: (() => void) | undefined;

  onMount(() => {
    unsubscribe = onAuthStateChanged(auth, (user) => {
      if (user === null) {
        navigate('/signin', { replace: true });
      }
    });
  });

  onCleanup(() => {
    unsubscribe?.();
  });

  return <Outlet />;
};

export default ProtectedRoutes;
