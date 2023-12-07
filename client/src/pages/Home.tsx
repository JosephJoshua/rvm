import axios from 'axios';
import { auth } from '../lib/firebase';
import { Show, createEffect, createSignal, onCleanup, onMount } from 'solid-js';
import { User, onAuthStateChanged, signOut } from 'firebase/auth';

const Home = () => {
  const [user, setUser] = createSignal<User | undefined>(undefined);
  const [points, setPoints] = createSignal(0);

  const getPoints = async (user: User) => {
    const idToken = await user.getIdToken();
    const url = new URL('/users/points', import.meta.env.VITE_BACKEND_URL);

    const response = await axios.get(url.href, {
      headers: {
        Authorization: `Bearer ${idToken}`,
      },
    });

    const points = response.data;
    setPoints(points);
  };

  const handleLogout = () => {
    signOut(auth);
  };

  createEffect(() => {
    const u = user();
    if (u === undefined) return;

    getPoints(u);
  });

  let unsubscribe: (() => void) | undefined;

  onMount(() => {
    unsubscribe = onAuthStateChanged(auth, (user) => {
      setUser(user ?? undefined);
    });
  });

  onCleanup(() => {
    unsubscribe?.();
  });

  return (
    <div class="flex flex-col items-center justify-center py-16">
      <Show when={user() !== undefined} fallback={'Loading...'}>
        <div class="text-2xl font-medium">{user()?.displayName}</div>
        <div class="text-xl font-medium mt-2">{user()?.email}</div>
        <div class="text-xl font-medium mt-4 text-green-600">
          {points()} points
        </div>

        <div class="flex flex-col gap-2 items-stretch mt-12">
          <a
            href="/scan"
            class="text-center transition duration-300 bg-green-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
          >
            Scan
          </a>

          <button
            type="button"
            class="text-center transition duration-300 bg-red-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
            onClick={handleLogout}
          >
            Log out
          </button>
        </div>
      </Show>
    </div>
  );
};

export default Home;
