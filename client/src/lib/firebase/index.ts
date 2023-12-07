import { initializeApp } from 'firebase/app';
import { getAuth } from 'firebase/auth';

const firebaseConfig = {
  apiKey: 'AIzaSyDy8TdQVBIiYHcvkDQu6bDzNvClKHzosNM',
  authDomain: 'greenwaste-rvm.firebaseapp.com',
  projectId: 'greenwaste-rvm',
  storageBucket: 'greenwaste-rvm.appspot.com',
  messagingSenderId: '865037019896',
  appId: '1:865037019896:web:c5e0813e8dee2df099690d',
  measurementId: 'G-BW512N2HKG',
};

const app = initializeApp(firebaseConfig);

export const auth = getAuth(app);

export default app;
