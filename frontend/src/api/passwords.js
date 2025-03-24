import axios from 'axios';

// 获取所有密码
export async function getAllPasswords() {
  try {
    const response = await axios.get('/api/passwords');
    console.log('密码API返回原始数据:', response);
    return response.data;
  } catch (error) {
    console.error('API错误 - 获取密码列表:', error);
    throw error;
  }
}

// 创建新密码
export async function createPassword(passwordData) {
  try {
    const response = await axios.post('/api/passwords', passwordData);
    return response.data;
  } catch (error) {
    console.error('API错误 - 创建密码:', error);
    throw error;
  }
}

// 更新密码
export async function updatePassword(passwordId, passwordData) {
  try {
    // 确保ID是一个有效的数字
    const id = Number(passwordId);
    if (isNaN(id) || id <= 0) {
      throw new Error(`无效的密码ID: ${passwordId}`);
    }
    
    // 确保数据对象不包含 ID 字段
    const { id: _, ...cleanData } = passwordData.id ? passwordData : { ...passwordData };
    
    const response = await axios.put(`/api/passwords/${id}`, cleanData);
    return response.data;
  } catch (error) {
    console.error(`更新密码失败:`, error);
    throw error;
  }
}

// 删除密码
export async function deletePassword(passwordId) {
  try {
    // 确保ID是一个有效的数字
    const id = Number(passwordId);
    if (isNaN(id)) {
      throw new Error('无效的密码ID');
    }
    
    const response = await axios.delete(`/api/passwords/${id}`);
    return response.data;
  } catch (error) {
    console.error('API错误 - 删除密码:', error);
    throw error;
  }
}

export default {
  getAllPasswords,
  createPassword,
  updatePassword,
  deletePassword
}; 