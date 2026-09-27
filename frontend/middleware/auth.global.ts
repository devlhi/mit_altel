// Apply the session guard to every page before rendering the workspace.
export default defineNuxtRouteMiddleware((to) => {
  const token = import.meta.client ? localStorage.getItem('isp_token') : null
  const isPublic = to.path === '/login'
  if (!token && !isPublic) {
    return navigateTo('/login')
  }
  if (token && isPublic) {
    return navigateTo('/')
  }
})
