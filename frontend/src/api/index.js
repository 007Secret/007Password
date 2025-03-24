import axios from 'axios';

// API基础URL - 默认为/api，但如果有环境变量，则使用环境变量
// 注意：移除所有可能包含的引号，避免URL编码问题
const API_BASE_URL = (typeof process !== 'undefined' && process.env && process.env.VUE_APP_API_URL) 
  ? process.env.VUE_APP_API_URL.replace(/["']/g, '') 
  : '/api';

console.log('使用API基础URL:', API_BASE_URL);

// 创建自定义axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  }
});

// 添加请求拦截器
api.interceptors.request.use(
  config => {
    // 从localStorage获取token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  response => response,
  error => {
    // 处理错误
    if (error.response && error.response.status === 401) {
      // 如果是未授权，清除token并记录日志
      console.log('未授权，需要重新登录');
      localStorage.removeItem('token');
      // 这里不主动跳转，让组件自己处理登录状态
    }
    return Promise.reject(error);
  }
);

// 如果token存在，设置默认headers
const token = localStorage.getItem('token');
if (token) {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
}

// API接口
export const auth = {
  // 登录
  login: async (masterPassword) => {
    try {
      const response = await api.post('/auth/login', { masterPassword });
      return response.data;
    } catch (error) {
      console.error('登录失败:', error);
      throw error;
    }
  },
  
  // 设置主密码
  setupMasterPassword: async (masterPassword) => {
    try {
      const response = await api.post('/auth/setup', { masterPassword });
      return response.data;
    } catch (error) {
      console.error('设置主密码失败:', error);
      throw error;
    }
  },
  
  // 检查首次使用
  checkFirstTimeSetup: async () => {
    try {
      const response = await api.get('/auth/check-first-time');
      console.log('API响应 - 首次使用检查:', response);
      return response.data;
    } catch (error) {
      console.error('检查首次使用状态失败:', error);
      throw error;
    }
  },
  
  // 验证token
  validate: async () => {
    try {
      const response = await api.get('/auth/validate');
      console.log('Token验证响应:', response);
      
      // 确保返回一个标准格式的响应，包括成功状态
      return {
        valid: true,
        data: response.data || {}
      };
    } catch (error) {
      console.error('验证token失败:', error);
      // 确保返回一个标准格式，即使是失败情况
      return {
        valid: false,
        error: error.response?.data?.error || '验证失败'
      };
    }
  },
  
  // 修改主密码
  changePassword: async (currentPassword, newPassword) => {
    try {
      const response = await api.post('/auth/change-password', { 
        currentPassword, 
        newPassword 
      });
      return response.data;
    } catch (error) {
      console.error('修改主密码失败:', error);
      throw error;
    }
  }
};

// 密码管理API
export const passwords = {
  // 获取所有密码
  getAllPasswords: async () => {
    try {
      const response = await api.get('/passwords');
      console.log('获取密码列表API响应:', response);
      // 确保返回一个对象，其中包含data字段，如果data不是数组则转换为空数组
      return {
        data: Array.isArray(response.data) ? response.data : 
              (response.data && typeof response.data === 'object') ? response.data : []
      };
    } catch (error) {
      console.error('获取密码列表失败:', error);
      // 即使发生错误，也返回一个有效的对象，其中data为空数组
      return { data: [] };
    }
  },
  
  // 获取单个密码
  getPassword: async (id) => {
    try {
      const response = await api.get(`/passwords/${id}`);
      return response.data;
    } catch (error) {
      console.error(`获取密码ID=${id}失败:`, error);
      throw error;
    }
  },
  
  // 创建新密码
  createPassword: async (passwordData) => {
    try {
      const response = await api.post('/passwords', passwordData);
      return response.data;
    } catch (error) {
      console.error('创建密码失败:', error);
      throw error;
    }
  },
  
  // 更新密码
  updatePassword: async (id, passwordData) => {
    try {
      const response = await api.put(`/passwords/${id}`, passwordData);
      return response.data;
    } catch (error) {
      console.error(`更新密码ID=${id}失败:`, error);
      throw error;
    }
  },
  
  // 删除密码
  deletePassword: async (id) => {
    try {
      const response = await api.delete(`/passwords/${id}`);
      return response.data;
    } catch (error) {
      console.error(`删除密码ID=${id}失败:`, error);
      throw error;
    }
  },
  
  // 搜索密码
  searchPasswords: async (query) => {
    try {
      const response = await api.get(`/passwords/search?q=${query}`);
      return response.data;
    } catch (error) {
      console.error('搜索密码失败:', error);
      throw error;
    }
  },
  
  // 导出密码为CSV
  exportToCSV: (passwords) => {
    try {
      // CSV 表头
      const headers = ['name', 'username', 'password', 'email', 'phone', 'website', 'notes'];
      
      // 创建CSV内容
      let csvContent = headers.join(',') + '\n';
      
      // 添加每行数据
      passwords.forEach(pwd => {
        const row = [
          // 转义字段，处理逗号和引号
          escapeCsvField(pwd.name || ''),
          escapeCsvField(pwd.username || ''),
          escapeCsvField(pwd.password || ''),
          escapeCsvField(pwd.email || ''),
          escapeCsvField(pwd.phone || ''),
          escapeCsvField(pwd.website || ''),
          escapeCsvField(pwd.notes || '')
        ];
        csvContent += row.join(',') + '\n';
      });
      
      return csvContent;
    } catch (error) {
      console.error('生成CSV失败:', error);
      throw error;
    }
  }
};

// CSV字段转义辅助函数
function escapeCsvField(field) {
  // 如果字段包含逗号、双引号或换行符，需要用双引号包裹并处理内部的双引号
  if (field.includes(',') || field.includes('"') || field.includes('\n')) {
    return '"' + field.replace(/"/g, '""') + '"';
  }
  return field;
}

export default {
  auth,
  passwords
}; 