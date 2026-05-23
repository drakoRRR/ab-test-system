import { initializeApp, getApps } from 'firebase/app'
import {
  getAuth,
  GoogleAuthProvider,
  signInWithPopup,
  signOut as firebaseSignOut,
  onIdTokenChanged,
  type User,
} from 'firebase/auth'

const firebaseConfig = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY ?? 'placeholder',
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN ?? '',
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID ?? '',
  storageBucket: process.env.NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET ?? '',
  messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID ?? '',
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID ?? '',
}

// Lazy init: getApps() avoids double-init in hot-reload and SSR
const app = getApps().length === 0 ? initializeApp(firebaseConfig) : getApps()[0]
export const auth = getAuth(app)

const TOKEN_COOKIE = 'firebase-token'

export function startTokenSync() {
  return onIdTokenChanged(auth, async (user) => {
    if (user) {
      const token = await user.getIdToken()
      document.cookie = `${TOKEN_COOKIE}=${token}; path=/; SameSite=Strict`
    } else {
      document.cookie = `${TOKEN_COOKIE}=; path=/; max-age=0`
    }
  })
}

export async function signInWithGoogle(): Promise<User> {
  const provider = new GoogleAuthProvider()
  const result = await signInWithPopup(auth, provider)
  return result.user
}

export async function signOut(): Promise<void> {
  await firebaseSignOut(auth)
}

export async function getIdToken(): Promise<string | null> {
  await auth.authStateReady()
  return auth.currentUser?.getIdToken() ?? null
}
