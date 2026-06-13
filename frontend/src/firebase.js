import { initializeApp } from "firebase/app"
import { getAuth, GoogleAuthProvider, signInWithPopup, signOut } from "firebase/auth"

const firebaseConfig = {
  apiKey: "AIzaSyCAYhwqNpEyeqK3fUQWy5bJykKJFMaslQw",
  authDomain: "cinema-booking-8749c.firebaseapp.com",
  projectId: "cinema-booking-8749c",
  storageBucket: "cinema-booking-8749c.firebasestorage.app",
  messagingSenderId: "156148471969",
  appId: "1:156148471969:web:a908ba759095faf0c478e0"
}

const app = initializeApp(firebaseConfig)
export const auth = getAuth(app)
export const provider = new GoogleAuthProvider()

export const loginWithGoogle = async () => {
  const result = await signInWithPopup(auth, provider)
  const token = await result.user.getIdToken()
  return { user: result.user, token }
}

export const logout = async () => {
  await signOut(auth)
}