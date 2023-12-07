import {
  GoogleAuthProvider,
  signInWithPopup,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  getIdToken,
} from 'firebase/auth';
import axios, { AxiosError } from 'axios';
import { auth } from '../firebase';

const googleProvider = new GoogleAuthProvider();

export const isSignedIn = () => {
  return auth.currentUser !== null;
};

export const registerUser = async () => {
  if (auth.currentUser === null) throw new Error('user not signed in');

  const idToken = await getIdToken(auth.currentUser);

  const url = new URL('/auth/register', import.meta.env.VITE_BACKEND_URL);

  return axios
    .postForm(url.href, {
      id_token: idToken,
    })
    .catch((err: AxiosError) => {
      if (err.status === 409) return;
      throw err;
    });
};

export const signInWithGoogle = async () => {
  return signInWithPopup(auth, googleProvider);
};

export const signInWithPassword = async (email: string, password: string) => {
  return signInWithEmailAndPassword(auth, email, password);
};

export const signUpWithPassword = async (email: string, password: string) => {
  return createUserWithEmailAndPassword(auth, email, password);
};
