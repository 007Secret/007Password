import axios from 'axios';

// 登录API，提交主密码并获取token
export async function login(masterPassword) {
  try {
    const response = await axios.post('/api/auth/login', { masterPassword });
    return response.data;
  } catch (error) {
    console.error('登录API错误:', error);
    throw error;
  }
}

// 验证token是否有效
export async function validate() {
  try {
    const response = await axios.get('/api/auth/validate');
    return response.data;
  } catch (error) {
    console.error('Token验证错误:', error);
    return { valid: false, error: error.message };
  }
}

// 检查是否首次使用需要设置主密码
export async function checkFirstTimeSetup() {
  try {
    const response = await axios.get('/api/auth/check-first-time');
    return response.data;
  } catch (error) {
    console.error('检查首次设置状态错误:', error);
    throw error;
  }
}

// 首次使用设置主密码
export async function setupMasterPassword(masterPassword) {
  try {
    const response = await axios.post('/api/auth/setup', { masterPassword });
    return response.data;
  } catch (error) {
    console.error('设置主密码错误:', error);
    throw error;
  }
}

// 修改主密码
export async function changePassword(currentPassword, newPassword) {
  try {
    const response = await axios.post('/api/auth/change-password', {
      currentPassword,
      newPassword
    });
    return response.data;
  } catch (error) {
    console.error('修改主密码错误:', error);
    throw error;
  }
}

// 备份所有密码数据（修改主密码前）
export async function backupPasswords() {
  try {
    const response = await axios.post('/api/auth/backup');
    return response.data;
  } catch (error) {
    console.error('备份密码数据错误:', error);
    throw error;
  }
}

// 恢复密码数据（如果修改主密码失败）
export async function restoreFromBackup() {
  try {
    const response = await axios.post('/api/auth/restore');
    return response.data;
  } catch (error) {
    console.error('恢复备份数据错误:', error);
    throw error;
  }
}

export default {
  login,
  validate,
  checkFirstTimeSetup,
  setupMasterPassword,
  changePassword,
  backupPasswords,
  restoreFromBackup
}; 