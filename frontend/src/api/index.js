import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

// Request interceptor: inject token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor: handle 401
api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(err)
  }
)

export default api

// Auth
export const authApi = {
  login: (data) => api.post('/auth/login', data),
  profile: () => api.get('/auth/profile'),
  changePassword: (data) => api.put('/auth/password', data)
}

// Projects
export const projectApi = {
  list: (params) => api.get('/projects', { params }),
  get: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post('/projects', data),
  update: (id, data) => api.put(`/projects/${id}`, data),
  delete: (id) => api.delete(`/projects/${id}`),
  restore: (id) => api.post(`/projects/${id}/restore`),
  listMembers: (id) => api.get(`/projects/${id}/members`),
  addMember: (id, data) => api.post(`/projects/${id}/members`, data),
  removeMember: (id, uid) => api.delete(`/projects/${id}/members/${uid}`)
}

// Documents
export const docApi = {
  list: (pid) => api.get(`/projects/${pid}/docs`),
  get: (pid, did) => api.get(`/projects/${pid}/docs/${did}`),
  create: (pid, data) => api.post(`/projects/${pid}/docs`, data),
  update: (pid, did, data) => api.put(`/projects/${pid}/docs/${did}`, data),
  delete: (pid, did) => api.delete(`/projects/${pid}/docs/${did}`)
}

// Users
export const userApi = {
  list: () => api.get('/users'),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`)
}

// Search
export const searchApi = {
  search: (q) => api.get('/search', { params: { q } })
}

// Settings
export const settingsApi = {
  get: () => api.get('/settings'),
  update: (data) => api.put('/settings', data)
}

// API Keys
export const apiKeyApi = {
  list: () => api.get('/apikeys'),
  create: (data) => api.post('/apikeys', data),
  delete: (id) => api.delete(`/apikeys/${id}`)
}

// Update
export const updateApi = {
  check: () => api.get('/update/check'),
  apply: () => api.post('/update/apply'),
  version: () => api.get('/version')
}

// File upload
export const fileApi = {
  upload: (formData, onProgress) => api.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress
  })
}
