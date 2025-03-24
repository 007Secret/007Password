<template>
  <div class="min-h-screen bg-gray-50">
    <!-- 登录表单 -->
    <div v-if="showLoginForm" class="flex items-center justify-center min-h-screen">
      <div class="w-full max-w-md p-8 bg-white rounded-lg shadow-md">
        <div class="flex justify-center mb-8">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <h2 class="mb-6 text-3xl font-bold text-center text-gray-800">007Password</h2>
        <div v-if="loginError" class="mb-4 text-sm text-center text-red-600">{{ loginError }}</div>
        
        <!-- 首次使用设置主密码 -->
        <div v-if="isFirstTimeSetup" class="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
          <p class="text-sm text-yellow-700 mb-2">欢迎使用007Password！这是您首次使用，请设置一个安全的主密码。</p>
          <p class="text-xs text-yellow-600">
            <strong>重要提示：</strong>主密码将用于加密您的所有密码数据。请务必记住您的主密码，如果忘记将无法恢复您的数据。
          </p>
        </div>
        
        <form @submit.prevent="handleLogin">
          <div class="mb-6">
            <label for="masterPassword" class="block mb-2 text-sm font-medium text-gray-700">
              {{ isFirstTimeSetup ? '设置主密码' : '主密码' }}
            </label>
            <input
              id="masterPassword"
              v-model="masterPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              :placeholder="isFirstTimeSetup ? '请设置您的主密码' : '请输入您的主密码'"
              autocomplete="current-password"
            />
          </div>
          
          <!-- 首次使用需要确认密码 -->
          <div v-if="isFirstTimeSetup" class="mb-6">
            <label for="confirmPassword" class="block mb-2 text-sm font-medium text-gray-700">确认主密码</label>
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="请再次输入主密码"
            />
          </div>
          
          <button type="submit" class="w-full px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2" :disabled="isLoading">
            {{ isLoading ? '登录中...' : (isFirstTimeSetup ? '设置主密码并登录' : '登录') }}
          </button>
        </form>
      </div>
    </div>

    <!-- 密码管理界面 -->
    <div v-else>
      <header class="sticky top-0 z-10 flex items-center justify-between w-full px-4 py-2 text-white bg-blue-700 shadow-md h-14">
        <h1 class="text-xl font-bold ml-12">007Password</h1>
        <div class="flex items-center mr-12">
          <div v-if="isLoggedIn" class="flex items-center space-x-6">
            <button @click="showChangePasswordModal = true" class="px-4 py-1 text-sm font-medium bg-blue-600 rounded-md hover:bg-blue-800">
              修改主密码
            </button>
            <button @click="logout" class="px-4 py-1 text-sm font-medium bg-red-600 rounded-md hover:bg-red-700">
              退出
            </button>
          </div>
        </div>
      </header>

      <main class="container px-4 py-6 mx-auto">
        <!-- 搜索和添加按钮 -->
        <div class="flex flex-col mb-6 space-y-4 md:flex-row md:space-y-0 md:space-x-4 md:items-center">
          <div class="relative flex-1">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索密码..."
              class="w-full px-4 py-2 pl-10 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <svg xmlns="http://www.w3.org/2000/svg" class="absolute w-5 h-5 text-gray-400 left-3 top-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <div class="flex space-x-3">
            <button @click="openAddModal" class="flex items-center px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              添加密码
            </button>
            <button @click="triggerImportFile" class="flex items-center px-4 py-2 text-white bg-green-600 rounded-md hover:bg-green-700">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
              </svg>
              导入CSV
            </button>
            <button @click="exportPasswordsToCSV" class="flex items-center px-4 py-2 text-yellow-900 bg-yellow-300 rounded-md hover:bg-yellow-400">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3M3 17v3a2 2 0 002 2h14a2 2 0 002-2v-3" />
              </svg>
              导出CSV
            </button>
            <input 
              type="file" 
              ref="fileInput" 
              accept=".csv" 
              class="hidden" 
              @change="handleFileImport"
            />
          </div>
        </div>

        <!-- 密码列表 -->
        <div class="overflow-hidden bg-white rounded-lg shadow">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">ID</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">应用名称</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">用户名</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">手机号</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">邮箱</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">密码</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">网站地址</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">授权登录</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">备注</th>
                <th scope="col" class="px-4 py-3 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="!filteredPasswords || filteredPasswords.length === 0">
                <td colspan="10" class="px-4 py-4 text-center text-gray-500">暂无数据</td>
              </tr>
              <tr v-for="(password, index) in filteredPasswords || []" :key="password.id" class="hover:bg-gray-50">
                <td class="px-4 py-4 whitespace-nowrap text-gray-500">{{ password.id }}</td>
                <td class="px-4 py-4 whitespace-nowrap">{{ password.name }}</td>
                <td class="px-4 py-4 whitespace-nowrap">{{ password.username || '-' }}</td>
                <td class="px-4 py-4 whitespace-nowrap">{{ password.phone || '-' }}</td>
                <td class="px-4 py-4 whitespace-nowrap">{{ password.email || '-' }}</td>
                <td class="px-4 py-4 whitespace-nowrap">
                  <div class="flex items-center">
                    <span>{{ password.showPassword ? password.password : '•••••••••••' }}</span>
                    <button 
                      @click="togglePasswordVisibility(index)" 
                      class="ml-2 text-gray-500 hover:text-blue-500"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path 
                          stroke-linecap="round" 
                          stroke-linejoin="round" 
                          stroke-width="2" 
                          :d="password.showPassword 
                            ? 'M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.542 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21' 
                            : 'M15 12a3 3 0 11-6 0 3 3 0 016 0z M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z'"
                        />
                      </svg>
                    </button>
                  </div>
                </td>
                <td class="px-4 py-4 whitespace-nowrap">
                  <a v-if="password.website" :href="addHttpToUrl(password.website)" target="_blank" class="text-blue-600 hover:text-blue-800">
                    {{ formatWebsite(password.website) }}
                  </a>
                  <span v-else>-</span>
                </td>
                <td class="px-4 py-4">
                  <div class="flex flex-wrap gap-1">
                    <!-- Auth logins 图标 -->
                    <div v-for="(enabled, provider) in password.authLogins || {}" :key="provider">
                      <div v-if="enabled" class="p-1 rounded-full" :title="provider">
                        <!-- Replace component with simple text representation -->
                        <span class="inline-block w-4 h-4 text-xs text-center bg-gray-200 rounded-full">
                          {{ provider.charAt(0).toUpperCase() }}
                        </span>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-4 whitespace-nowrap">
                  <span v-if="password.notes">{{ password.notes }}</span>
                  <span v-else>-</span>
                </td>
                <td class="px-4 py-4 whitespace-nowrap">
                  <div class="flex items-center justify-end space-x-2">
                    <button @click="openViewModal(password)" class="p-1 text-gray-600 hover:text-gray-800">
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    </button>
                    <button @click="openEditModal(password)" class="p-1 text-blue-600 hover:text-blue-800">
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>
                    <button @click="confirmDeletePassword(password.id)" class="p-1 text-red-600 hover:text-red-800">
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </main>

      <!-- 添加/编辑密码表单 -->
      <div v-if="showModal" class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="closeModal"></div>
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <h3 class="mb-4 text-lg font-medium text-gray-900">{{ isEditing ? '编辑密码' : '添加密码' }}</h3>
            <form @submit.prevent="submitForm">
              <div class="grid grid-cols-1 gap-4">
                <div>
                  <label for="passwordName" class="block mb-2 text-sm font-medium text-gray-700">应用名称 <span class="text-red-500">*</span></label>
                  <input
                    id="passwordName"
                    v-model="formData.name"
                    type="text"
                    class="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                  <div v-if="formErrors.name" class="mt-1 text-sm text-red-600">{{ formErrors.name }}</div>
                </div>
                <div>
                  <label for="username" class="block text-sm font-medium text-gray-700">用户名</label>
                  <input
                    id="username"
                    v-model="formData.username"
                    type="text"
                    class="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label for="phone" class="block text-sm font-medium text-gray-700">手机号</label>
                  <input
                    id="phone"
                    v-model="formData.phone"
                    type="text"
                    class="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label for="email" class="block text-sm font-medium text-gray-700">邮箱</label>
                  <input
                    id="email"
                    v-model="formData.email"
                    type="email"
                    class="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label for="password" class="block text-sm font-medium text-gray-700">密码</label>
                  <div class="relative mt-1">
                    <input
                      id="password"
                      v-model="formData.password"
                      :type="showPassword ? 'text' : 'password'"
                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <button
                      type="button"
                      @click="showPassword = !showPassword"
                      class="absolute inset-y-0 right-0 flex items-center px-3 text-gray-500"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path
                          v-if="showPassword"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.542 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
                        />
                        <path
                          v-else
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                        />
                        <path
                          v-if="!showPassword"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                        />
                      </svg>
                    </button>
                  </div>
                </div>
                <div>
                  <label for="passwordWebsite" class="block mb-2 text-sm font-medium text-gray-700">网址</label>
                  <input
                    id="passwordWebsite"
                    v-model="formData.website"
                    type="text"
                    class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    placeholder="https://example.com"
                  />
                  <div v-if="formData.website" class="mt-1 text-xs text-gray-500">
                    显示效果: {{ formatWebsite(formData.website) }}
                  </div>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700">授权登录</label>
                  <div class="grid grid-cols-2 gap-2 mt-1">
                    <div>
                      <input
                        id="google"
                        v-model="selectedAuthLogins.google"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="google" class="ml-2 text-sm text-gray-700">谷歌</label>
                    </div>
                    <div>
                      <input
                        id="wechat"
                        v-model="selectedAuthLogins.wechat"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="wechat" class="ml-2 text-sm text-gray-700">微信</label>
                    </div>
                    <div>
                      <input
                        id="weibo"
                        v-model="selectedAuthLogins.weibo"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="weibo" class="ml-2 text-sm text-gray-700">微博</label>
                    </div>
                    <div>
                      <input
                        id="baidu"
                        v-model="selectedAuthLogins.baidu"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="baidu" class="ml-2 text-sm text-gray-700">百度</label>
                    </div>
                    <div>
                      <input
                        id="facebook"
                        v-model="selectedAuthLogins.facebook"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="facebook" class="ml-2 text-sm text-gray-700">Facebook</label>
                    </div>
                    <div>
                      <input
                        id="github"
                        v-model="selectedAuthLogins.github"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="github" class="ml-2 text-sm text-gray-700">Github</label>
                    </div>
                    <div>
                      <input
                        id="twitter"
                        v-model="selectedAuthLogins.twitter"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="twitter" class="ml-2 text-sm text-gray-700">X(Twitter)</label>
                    </div>
                    <div>
                      <input
                        id="qq"
                        v-model="selectedAuthLogins.qq"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="qq" class="ml-2 text-sm text-gray-700">QQ</label>
                    </div>
                    <div>
                      <input
                        id="alipay"
                        v-model="selectedAuthLogins.alipay"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="alipay" class="ml-2 text-sm text-gray-700">支付宝</label>
                    </div>
                    <div>
                      <input
                        id="taobao"
                        v-model="selectedAuthLogins.taobao"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="taobao" class="ml-2 text-sm text-gray-700">淘宝</label>
                    </div>
                    <div>
                      <input
                        id="dingtalk"
                        v-model="selectedAuthLogins.dingtalk"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="dingtalk" class="ml-2 text-sm text-gray-700">钉钉</label>
                    </div>
                    <div>
                      <input
                        id="douyin"
                        v-model="selectedAuthLogins.douyin"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="douyin" class="ml-2 text-sm text-gray-700">抖音</label>
                    </div>
                    <div>
                      <input
                        id="feishu"
                        v-model="selectedAuthLogins.feishu"
                        type="checkbox"
                        class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                      <label for="feishu" class="ml-2 text-sm text-gray-700">飞书</label>
                    </div>
                  </div>
                </div>
                <div>
                  <label for="notes" class="block text-sm font-medium text-gray-700">备注</label>
                  <textarea
                    id="notes"
                    v-model="formData.notes"
                    rows="3"
                    class="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  ></textarea>
                </div>
              </div>
              <div class="flex justify-end mt-6 space-x-3">
                <button type="button" @click="closeModal" class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300">
                  取消
                </button>
                <button type="submit" class="px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700">
                  保存
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- 查看密码弹窗 -->
      <div v-if="showViewModal" class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="closeViewModal"></div>
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <h3 class="mb-4 text-lg font-medium text-gray-900">密码详情</h3>
            <div class="space-y-4">
              <div>
                <h4 class="text-sm font-medium text-gray-700">应用名称</h4>
                <p class="mt-1 text-gray-900">{{ viewData.name }}</p>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-700">用户名</h4>
                <p class="mt-1 text-gray-900">{{ viewData.username || '-' }}</p>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-700">手机号</h4>
                <p class="mt-1 text-gray-900">{{ viewData.phone || '-' }}</p>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-700">邮箱</h4>
                <p class="mt-1 text-gray-900">{{ viewData.email || '-' }}</p>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-700">密码</h4>
                <div class="flex items-center mt-1">
                  <p class="text-gray-900">{{ showViewPassword ? viewData.password : '••••••••' }}</p>
                  <button @click="showViewPassword = !showViewPassword" class="ml-2 text-blue-600">
                    {{ showViewPassword ? '隐藏' : '显示' }}
                  </button>
                </div>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-700">网站地址</h4>
                <p class="mt-1">
                  <a v-if="viewData.website" :href="addHttpToUrl(viewData.website)" target="_blank" class="text-blue-600 hover:text-blue-800">
                    {{ formatWebsite(viewData.website) }}
                  </a>
                  <span v-else class="text-gray-500">-</span>
                </p>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-700">授权登录</h4>
                <div class="flex flex-wrap gap-2 mt-1">
                  <span v-if="viewData.authLogins?.google" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">谷歌</span>
                  <span v-if="viewData.authLogins?.wechat" class="px-2 py-1 text-xs bg-green-100 text-green-800 rounded-full">微信</span>
                  <span v-if="viewData.authLogins?.weibo" class="px-2 py-1 text-xs bg-red-100 text-red-800 rounded-full">微博</span>
                  <span v-if="viewData.authLogins?.baidu" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">百度</span>
                  <span v-if="viewData.authLogins?.facebook" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">Facebook</span>
                  <span v-if="viewData.authLogins?.github" class="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded-full">Github</span>
                  <span v-if="viewData.authLogins?.twitter" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">X(Twitter)</span>
                  <span v-if="viewData.authLogins?.qq" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">QQ</span>
                  <span v-if="viewData.authLogins?.alipay" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">支付宝</span>
                  <span v-if="viewData.authLogins?.taobao" class="px-2 py-1 text-xs bg-orange-100 text-orange-800 rounded-full">淘宝</span>
                  <span v-if="viewData.authLogins?.dingtalk" class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">钉钉</span>
                  <span v-if="viewData.authLogins?.douyin" class="px-2 py-1 text-xs bg-black text-white rounded-full">抖音</span>
                  <span v-if="viewData.authLogins?.feishu" class="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded-full">飞书</span>
                  <span v-if="!hasAuthLogins(viewData.authLogins)" class="text-gray-500">-</span>
                </div>
              </div>
              <div>
                <h4 class="text-sm font-medium text-gray-700">备注</h4>
                <p class="mt-1 text-gray-900 whitespace-pre-line">{{ viewData.notes || '-' }}</p>
              </div>
            </div>
            <div class="flex justify-end mt-6">
              <button @click="closeViewModal" class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300">
                关闭
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 删除确认弹窗 -->
      <div v-if="showDeleteModal" class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="closeDeleteModal"></div>
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <h3 class="mb-4 text-lg font-medium text-gray-900">确认删除</h3>
            <p class="text-gray-600">您确定要删除 "{{ passwordToDelete?.name }}" 吗？此操作无法撤销。</p>
            <div class="flex justify-end mt-6 space-x-3">
              <button @click="closeDeleteModal" class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300">
                取消
              </button>
              <button @click="deletePassword" class="px-4 py-2 text-white bg-red-600 rounded-md hover:bg-red-700">
                删除
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 修改主密码弹窗 -->
      <div v-if="showChangePasswordModal" class="fixed inset-0 z-20 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="showChangePasswordModal = false"></div>
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <h3 class="mb-4 text-lg font-medium text-gray-900">修改主密码</h3>
            <form @submit.prevent="changeMasterPassword">
              <div class="mb-4">
                <label for="currentPassword" class="block mb-2 text-sm font-medium text-gray-700">
                  当前主密码
                </label>
                <input
                  id="currentPassword"
                  v-model="passwordForm.currentPassword"
                  type="password"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="请输入当前主密码"
                />
              </div>
              
              <div class="mb-4">
                <label for="newPassword" class="block mb-2 text-sm font-medium text-gray-700">
                  新主密码
                </label>
                <input
                  id="newPassword"
                  v-model="passwordForm.newPassword"
                  type="password"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="请输入新主密码（至少6位）"
                />
              </div>
              
              <div class="mb-6">
                <label for="confirmNewPassword" class="block mb-2 text-sm font-medium text-gray-700">
                  确认新主密码
                </label>
                <input
                  id="confirmNewPassword"
                  v-model="passwordForm.confirmNewPassword"
                  type="password"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="请再次输入新主密码"
                />
              </div>
              
              <div v-if="passwordChangeError" class="mb-4 p-2 text-sm text-center text-red-600 bg-red-50 border border-red-200 rounded">
                {{ passwordChangeError }}
              </div>
              
              <div class="flex justify-end space-x-3">
                <button 
                  type="button" 
                  @click="showChangePasswordModal = false" 
                  class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300"
                >
                  取消
                </button>
                <button 
                  type="submit" 
                  class="px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700"
                  :disabled="isChangingPassword"
                >
                  {{ isChangingPassword ? '处理中...' : '保存' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- 主密码修改成功弹窗 -->
      <div v-if="showSuccessModal && successType === 'passwordChange'" class="fixed inset-0 z-20 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="showSuccessModal = false"></div>
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <div class="flex items-center justify-center mb-4">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="mb-4 text-xl font-medium text-center text-gray-900">主密码修改成功</h3>
            <p class="mb-5 text-gray-700 text-center">
              您的主密码已成功修改！<br>
              系统将退出登录，请使用新密码重新登录。<br><br>
              <strong class="text-blue-600">重要提示：</strong><br>
              主密码用于加密所有密码数据，请务必牢记您的新主密码。
            </p>
            <div class="flex justify-center mt-6">
              <button 
                type="button" 
                @click="handlePasswordChangeSuccess" 
                class="px-6 py-2 font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              >
                确认并重新登录
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 密码添加成功弹窗 -->
      <div v-if="showSuccessModal && successType === 'passwordAdd'" class="fixed inset-0 z-20 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="showSuccessModal = false"></div>
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <div class="flex items-center justify-center mb-4">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="mb-4 text-xl font-medium text-center text-gray-900">密码添加成功</h3>
            <p class="mb-5 text-gray-700 text-center">
              您的密码已成功添加到密码管理器中！
            </p>
            <div class="flex justify-center mt-6">
              <button 
                type="button" 
                @click="showSuccessModal = false" 
                class="px-6 py-2 font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              >
                确认
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 导入CSV过程弹窗 -->
      <div v-if="showImportModal" class="fixed inset-0 z-20 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="closeImportModal"></div>
          <div class="relative w-full max-w-lg p-6 mx-auto bg-white rounded-lg shadow-xl">
            <h3 class="mb-4 text-lg font-medium text-gray-900">导入Chrome密码</h3>
            
            <div v-if="importStep === 'preview'" class="mb-4">
              <p class="mb-4 text-sm text-gray-600">已检测到 <strong>{{ importedPasswords.length }}</strong> 条密码记录。请确认是否导入：</p>
              
              <div class="max-h-64 overflow-y-auto border border-gray-200 rounded-md mb-4">
                <table class="min-w-full divide-y divide-gray-200">
                  <thead class="bg-gray-50">
                    <tr>
                      <th scope="col" class="px-3 py-2 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">网站</th>
                      <th scope="col" class="px-3 py-2 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">用户名</th>
                      <th scope="col" class="px-3 py-2 text-xs font-medium tracking-wider text-left text-gray-500 uppercase">密码</th>
                    </tr>
                  </thead>
                  <tbody class="bg-white divide-y divide-gray-200">
                    <tr v-for="(pwd, index) in importedPasswords.slice(0, 5)" :key="index" class="hover:bg-gray-50">
                      <td class="px-3 py-2 whitespace-nowrap">{{ pwd.name }}</td>
                      <td class="px-3 py-2 whitespace-nowrap">{{ pwd.username }}</td>
                      <td class="px-3 py-2 whitespace-nowrap">•••••••</td>
                    </tr>
                    <tr v-if="importedPasswords.length > 5">
                      <td colspan="3" class="px-3 py-2 text-center text-gray-500">
                        还有 {{ importedPasswords.length - 5 }} 条记录...
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              
              <div class="flex justify-end space-x-3">
                <button @click="closeImportModal" class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300">
                  取消
                </button>
                <button @click="confirmImport" class="px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700">
                  确认导入
                </button>
              </div>
            </div>
            
            <div v-if="importStep === 'processing'" class="mb-4">
              <div class="flex flex-col items-center">
                <svg class="animate-spin h-10 w-10 text-blue-600 mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <p class="text-sm text-gray-600">正在导入密码 ({{ importProgress }}/{{ importedPasswords.length }})...</p>
              </div>
            </div>
            
            <div v-if="importStep === 'complete'" class="mb-4">
              <div class="flex flex-col items-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-green-500 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p class="text-center text-gray-700">
                  导入完成！成功导入 <strong class="text-green-600">{{ importSuccessCount }}</strong> 条密码。
                  <span v-if="importFailedCount > 0" class="block mt-2 text-red-600">
                    {{ importFailedCount }} 条密码导入失败。
                  </span>
                </p>
              </div>
              
              <div class="flex justify-center mt-6">
                <button @click="closeImportModal" class="px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700">
                  完成
                </button>
              </div>
            </div>
            
            <div v-if="importStep === 'error'" class="mb-4">
              <div class="flex flex-col items-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-red-500 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p class="text-center text-gray-700">
                  导入失败：<span class="text-red-600">{{ importError }}</span>
                </p>
              </div>
              
              <div class="flex justify-center mt-6">
                <button @click="closeImportModal" class="px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700">
                  关闭
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 恢复备份确认弹窗 -->
      <div v-if="showRestoreBackupModal" class="fixed inset-0 z-20 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="showRestoreBackupModal = false"></div>
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <h3 class="mb-4 text-lg font-medium text-gray-900">恢复密码备份</h3>
            <p class="mb-4 text-gray-600">
              检测到有 <strong>{{ backupPasswordsCount }}</strong> 条密码备份数据，可能是您在修改主密码过程中生成的备份。
              您当前有 <strong>{{ passwordsList.length }}</strong> 条密码数据。
            </p>
            <p class="mb-6 text-yellow-600 font-medium">
              恢复备份操作会将当前密码数据替换为备份数据，确定要继续吗？
            </p>
            <div class="flex justify-end space-x-3">
              <button @click="showRestoreBackupModal = false" class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300">
                取消
              </button>
              <button @click="restoreBackupData" class="px-4 py-2 text-white bg-green-600 rounded-md hover:bg-green-700">
                确认恢复
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 认证登录图标编辑模态窗口 -->
      <div v-if="showAuthLoginsModal" class="fixed inset-0 z-20 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 transition-opacity bg-black bg-opacity-50" @click="showAuthLoginsModal = false"></div>
          
          <div class="relative w-full max-w-md p-6 mx-auto bg-white rounded-lg shadow-xl">
            <div class="mb-4">
              <h3 class="text-lg font-medium text-gray-900">编辑认证登录</h3>
            </div>
            
            <div class="space-y-4">
              <div v-for="(provider, key) in authProviders" :key="key" class="flex items-center justify-between">
                <div class="flex items-center">
                  <img :src="provider.icon" :alt="provider.name" class="w-8 h-8 mr-2" />
                  <span>{{ provider.name }}</span>
                </div>
                
                <div>
                  <label class="inline-flex items-center cursor-pointer">
                    <input 
                      type="checkbox" 
                      class="sr-only peer" 
                      :checked="editingAuthLogins[key]" 
                      @change="editingAuthLogins[key] = !editingAuthLogins[key]"
                    >
                    <div class="relative w-11 h-6 bg-gray-200 rounded-full peer peer-focus:ring-4 peer-focus:ring-blue-300 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                  </label>
                </div>
              </div>
            </div>
            
            <div class="mt-6 space-x-2 text-right">
              <button @click="showAuthLoginsModal = false" class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200">
                取消
              </button>
              <button @click="saveAuthLogins" class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700">
                保存
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, reactive, nextTick, watch } from 'vue';
import { auth, passwords } from '../api';
import axios from 'axios';
import { useRouter } from 'vue-router';
import { useMessage } from 'naive-ui';
import { markRaw, defineAsyncComponent } from 'vue';

// 登录表单相关
const masterPassword = ref('');
const loginError = ref('');
const isLoading = ref(false);
const isLoggedIn = ref(false);
const showLoginForm = ref(true);
const isFirstTimeSetup = ref(false);
const confirmPassword = ref('');

// 路由
const router = useRouter();
// 消息提示
const message = useMessage();

// 密码管理相关状态
const passwordsList = ref([]);
const searchQuery = ref('');
const showModal = ref(false);
const isEditing = ref(false);
const showPassword = ref(false);
const formData = ref(createEmptyForm());
const formError = ref('');
const editingPassword = ref(createEmptyForm());
const isSubmitting = ref(false); // Add the missing isSubmitting ref
const selectedAuthLogins = ref({
  google: false,
  wechat: false,
  weibo: false,
  baidu: false,
  facebook: false,
  github: false,
  qq: false,
  alipay: false,
  taobao: false,
  dingtalk: false,
  douyin: false,
  feishu: false,
  twitter: false
});
const showViewModal = ref(false);
const viewData = ref({
  name: '',
  username: '',
  phone: '',
  password: '',
  website: '',
  authLogins: {
    google: false,
    wechat: false,
    weibo: false,
    baidu: false,
    facebook: false,
    github: false,
    qq: false,
    alipay: false,
    taobao: false,
    dingtalk: false,
    douyin: false,
    feishu: false,
    twitter: false
  },
  notes: ''
});
const showViewPassword = ref(false);
const showDeleteModal = ref(false);
const passwordToDelete = ref(null);
const showChangePasswordModal = ref(false);
const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmNewPassword: ''
});
const passwordChangeError = ref('');
const isChangingPassword = ref(false);
const showSuccessModal = ref(false);
const successType = ref('passwordAdd');

// CSV导入相关状态
const showImportModal = ref(false);
const fileInput = ref(null);
const importedPasswords = ref([]);
const importStep = ref('preview'); // 'preview', 'processing', 'complete', 'error'
const importProgress = ref(0);
const importError = ref('');
const importSuccessCount = ref(0);
const importFailedCount = ref(0);

// 添加 showPasswordMap 变量
const showPasswordMap = ref({});

// 导出CSV功能
function exportPasswordsToCSV() {
  try {
    // 获取要导出的密码列表（可以使用当前过滤后的密码或全部密码）
    const passwordsToExport = filteredPasswords.value || [];
    
    if (passwordsToExport.length === 0) {
      message.warning('没有可导出的密码');
      return;
    }
    
    // 使用API中的方法生成CSV内容
    const csvContent = passwords.exportToCSV(passwordsToExport);
    
    // 创建Blob
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    
    // 创建URL
    const url = URL.createObjectURL(blob);
    
    // 创建下载链接
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', '007Password导出_' + new Date().toISOString().split('T')[0] + '.csv');
    link.style.visibility = 'hidden';
    
    // 添加到DOM并触发点击
    document.body.appendChild(link);
    link.click();
    
    // 清理
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
    
    // 显示成功消息
    message.success(`成功导出 ${passwordsToExport.length} 个密码条目`);
  } catch (error) {
    console.error('导出CSV失败:', error);
    message.error('导出CSV失败: ' + error.message);
  }
}

// 过滤后的密码列表
const filteredPasswords = computed(() => {
  if (!passwordsList.value || !Array.isArray(passwordsList.value)) {
    console.log('passwordsList不是数组或为空', passwordsList.value);
    return [];
  }
  
  // 没有搜索词，直接返回所有密码
  if (!searchQuery.value) {
    return passwordsList.value;
  }
  
  // 搜索词转为小写以进行不区分大小写的搜索
  const query = searchQuery.value.toLowerCase();
  
  // 过滤满足条件的密码
  return passwordsList.value.filter(password => {
    return (
      (password.name && password.name.toLowerCase().includes(query)) ||
      (password.username && password.username.toLowerCase().includes(query)) ||
      (password.phone && password.phone.toLowerCase().includes(query)) ||
      (password.website && password.website.toLowerCase().includes(query)) ||
      (password.notes && password.notes.toLowerCase().includes(query))
    );
  });
});

// 密码列表排序和处理函数
const sortedPasswords = computed(() => {
  // ... existing code ...
});

// 格式化网址显示，过长时折叠
function formatWebsite(website) {
  if (!website) return '';
  
  // 移除协议前缀
  let formattedUrl = website.replace(/^(https?:\/\/)?(www\.)?/i, '');
  
  // 如果网址超过25个字符，则截断并显示省略号
  if (formattedUrl.length > 25) {
    return formattedUrl.substring(0, 22) + '...';
  }
  
  return formattedUrl;
}

// API调用函数
const handleLogin = async () => {
  isLoading.value = true;
  loginError.value = '';
  
  try {
    // 先检查是否首次使用
    console.log('检查是否首次使用...');
    const checkResp = await auth.checkFirstTimeSetup();
    console.log('首次使用检查结果:', checkResp);
    
    // 更新首次使用状态
    isFirstTimeSetup.value = checkResp.isFirstTimeSetup;
    
    // 首次使用时，如果用户正在设置密码
    if (isFirstTimeSetup.value) {
      console.log('这是首次使用，设置主密码');
      // 验证两次密码是否一致
      if (masterPassword.value !== confirmPassword.value) {
        loginError.value = '两次输入的密码不一致';
        isLoading.value = false;
        return;
      }
      
      // 调用设置主密码API
      const setupResp = await auth.setupMasterPassword(masterPassword.value);
      console.log('设置主密码响应:', setupResp);
      
      if (setupResp.token) {
        localStorage.setItem('token', setupResp.token);
        axios.defaults.headers.common['Authorization'] = `Bearer ${setupResp.token}`;
        
        // 更新UI状态
        showLoginForm.value = false;
        isLoggedIn.value = true;
        masterPassword.value = '';
        
        // 获取密码列表
        await fetchPasswords();
        
        // 使用naive-ui显示成功消息
        message.success('主密码设置成功！');
      } else {
        loginError.value = '设置主密码失败';
      }
    } else {
      console.log('已经设置过主密码，进行登录');
      // 已经设置过主密码，调用登录API
      const loginResp = await auth.login(masterPassword.value);
      console.log('登录响应:', loginResp);
      
      if (loginResp.token) {
        localStorage.setItem('token', loginResp.token);
        axios.defaults.headers.common['Authorization'] = `Bearer ${loginResp.token}`;
        
        // 更新UI状态 - 确保这里正确设置了显示状态
        console.log('登录成功，设置状态为已登录并显示密码列表');
        showLoginForm.value = false;
        isLoggedIn.value = true;
        masterPassword.value = '';
        
        // 获取密码列表
        await fetchPasswords();
      } else {
        loginError.value = '登录失败，请检查主密码';
      }
    }
  } catch (error) {
    console.error('登录/设置过程发生错误:', error);
    loginError.value = error.response?.data?.error || '登录过程中发生错误';
  } finally {
    isLoading.value = false;
    // 再次检查登录状态，确保UI正确更新
    console.log('登录流程完成，当前登录状态:', isLoggedIn.value, '是否显示登录表单:', showLoginForm.value);
  }
};

function logout() {
  localStorage.removeItem('token');
  isLoggedIn.value = false;
  masterPassword.value = '';
}

async function fetchPasswords() {
  try {
    console.log('开始获取密码列表...');
    const response = await passwords.getAllPasswords();
    console.log('API返回数据结构:', response);
    
    // 检查响应格式，适配API返回的数据结构
    let passwordData = Array.isArray(response) ? response : 
                        (response.data && Array.isArray(response.data)) ? response.data : [];
    
    console.log('处理后的密码数据数量:', passwordData.length);
    
    // 确保每个密码都有 authLogins 对象
    passwordData = passwordData.map(pwd => ({
      ...pwd,
      authLogins: pwd.authLogins || {},
      showPassword: false // 添加显示密码标志
    }));
    
    // 初始化为空数组，确保总是有有效的数组
    passwordsList.value = passwordData;
    
    // 如果返回的数据不是数组或为空，则提前返回
    if (!Array.isArray(passwordsList.value) || passwordsList.value.length === 0) {
      console.log('获取到密码列表为空');
      return;
    }
    
    // 检查是否有解密失败的密码
    let hasDecryptionError = false;
    let errorCount = 0;
    
    passwordsList.value.forEach(pwd => {
      if (typeof pwd.password === 'string' && 
          (pwd.password.startsWith('解密失败:') || 
           pwd.password.includes('sql: no rows in result set'))) {
        hasDecryptionError = true;
        errorCount++;
        pwd.password = '解密失败: 主密码可能已更改';
      }
    });
    
    if (hasDecryptionError) {
      console.warn(`检测到 ${errorCount}/${passwordsList.value.length} 个密码解密失败`);
      
      // 如果所有密码都解密失败，可能是主密码已被修改
      if (errorCount === passwordsList.value.length && passwordsList.value.length > 0) {
        // 显示一个全局提示，建议用户重新登录
        const shouldRelogin = confirm(
          "检测到所有密码都无法解密，这通常是因为主密码已被修改。\n\n" +
          "建议您退出并使用正确的主密码重新登录。\n\n" +
          "是否立即退出登录？"
        );
        
        if (shouldRelogin) {
          logout();
          window.location.reload();
          return;
        }
      }
    }
    
    console.log('获取到密码列表，数量:', passwordsList.value.length);
  } catch (error) {
    console.error('获取密码列表失败:', error);
    // 确保即使出错也初始化为空数组
    passwordsList.value = [];
    if (error.response && error.response.status === 401) {
      // Token expired or invalid, redirect to login
      logout();
    }
  }
}

// 表单处理函数
function createEmptyForm() {
  return {
    id: null,
    name: '',
    username: '',
    phone: '',
    password: '',
    website: '',
    authLogins: {
      google: false,
      wechat: false,
      weibo: false,
      baidu: false,
      facebook: false,
      github: false,
      qq: false,
      alipay: false,
      taobao: false,
      dingtalk: false,
      douyin: false,
      feishu: false,
      twitter: false
    },
    notes: ''
  };
}

function openAddModal() {
  formData.value = {
    id: null,
    name: '',
    username: '',
    password: '',
    website: '',
    notes: '',
    phone: ''
  };
  resetAuthLogins();
  formErrors.value = {};
  isEditing.value = false;
  showModal.value = true;
}

function openEditModal(password) {
  // 创建全新对象，避免对象引用问题
  formData.value = {
    id: parseInt(password.id, 10), // 使用处理后的数字ID
    name: password.name || '',
    username: password.username || '',
    password: password.password || '',
    website: password.website || '',
    notes: password.notes || '',
    phone: password.phone || ''
  };
  
  mapAuthLoginsToForm(password.authLogins);
  formErrors.value = {};
  isEditing.value = true;
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
}

async function submitForm() {
  if (isEditing.value) {
    await updatePassword();
  } else {
    await addPassword();
  }
}

// 表单验证
function validateForm() {
  const errors = {};
  
  if (!formData.value.name.trim()) {
    errors.name = '应用名称不能为空';
  }
  
  formErrors.value = errors;
  return Object.keys(errors).length === 0;
}

// 添加密码
async function addPassword() {
  if (!validateForm()) {
    return;
  }
  
  try {
    isSubmitting.value = true;
    
    // 创建要提交的对象
    const passwordData = {
      name: formData.value.name,
      username: formData.value.username,
      password: formData.value.password,
      website: formData.value.website,
      notes: formData.value.notes,
      phone: formData.value.phone,
      authLogins: selectedAuthLogins.value
    };
    
    // 调用API保存密码
    await passwords.createPassword(passwordData);
    
    // 重置表单并关闭模态框
    resetForm();
    showModal.value = false;
    
    // 重新获取密码列表
    await fetchPasswords();
    
    // 显示成功提示
    showSuccessMessage('密码添加成功');
  } catch (error) {
    console.error('添加密码失败:', error);
    errorMessage.value = '添加密码失败: ' + (error.message || '未知错误');
  } finally {
    isSubmitting.value = false;
  }
}

// 编辑密码
async function updatePassword() {
  if (!validateForm()) {
    return;
  }
  
  try {
    isSubmitting.value = true;
    
    // 确保ID是有效的数字
    const idValue = parseInt(formData.value.id, 10);
    if (isNaN(idValue) || idValue <= 0) {
      throw new Error('密码ID无效，无法更新');
    }
    
    // 创建要提交的对象 - 手动构建以确保类型正确，删除ID字段
    const passwordData = {
      name: String(formData.value.name || ''),
      username: String(formData.value.username || ''),
      password: String(formData.value.password || ''),
      website: String(formData.value.website || ''),
      notes: String(formData.value.notes || ''),
      phone: String(formData.value.phone || ''),
      authLogins: { ...selectedAuthLogins.value } // 创建授权登录对象的副本
    };
    
    // 调用API更新密码 - 分别传递ID和数据
    await passwords.updatePassword(idValue, passwordData);
    
    // 重置表单并关闭模态框
    resetForm();
    showModal.value = false;
    
    // 显示成功提示
    message.success('密码更新成功');
    
    // 重新获取密码列表
    await fetchPasswords();
  } catch (error) {
    console.error('更新密码失败:', error);
    message.error('更新密码失败: ' + (error.message || '未知错误'));
  } finally {
    isSubmitting.value = false;
  }
}

// 查看密码
function viewPassword(password) {
  viewData.value = {
    id: password.id,
    name: password.name || '',
    username: password.username || '',
    phone: password.phone || '',
    password: password.password || '',
    website: password.website || '',
    authLogins: {
      google: password.authLogins?.google || false,
      wechat: password.authLogins?.wechat || false,
      weibo: password.authLogins?.weibo || false,
      baidu: password.authLogins?.baidu || false,
      facebook: password.authLogins?.facebook || false,
      github: password.authLogins?.github || false,
      qq: password.authLogins?.qq || false,
      alipay: password.authLogins?.alipay || false,
      taobao: password.authLogins?.taobao || false,
      dingtalk: password.authLogins?.dingtalk || false,
      douyin: password.authLogins?.douyin || false,
      feishu: password.authLogins?.feishu || false,
      twitter: password.authLogins?.twitter || false
    },
    notes: password.notes || ''
  };
  showViewPassword.value = false;
  showViewModal.value = true;
}

function closeViewModal() {
  showViewModal.value = false;
}

// 删除密码
function confirmDelete(password) {
  passwordToDelete.value = password;
  showDeleteModal.value = true;
}

function closeDeleteModal() {
  showDeleteModal.value = false;
  passwordToDelete.value = null;
}

async function deletePassword() {
  try {
    // 确保id是数字类型
    const id = Number(passwordToDelete.value.id);
    if (isNaN(id) || id <= 0) {
      throw new Error('无效的密码ID');
    }
    
    await passwords.deletePassword(id);
    closeDeleteModal();
    fetchPasswords();
  } catch (error) {
    console.error('删除密码失败:', error);
    // 显示错误提示
    message.error('删除密码失败: ' + error.message);
  }
}

// 辅助函数
function hasAuthLogins(authLogins) {
  if (!authLogins) return false;
  return Boolean(
    authLogins.google || authLogins.wechat || authLogins.weibo || authLogins.baidu ||
    authLogins.facebook || authLogins.github || authLogins.qq || authLogins.alipay ||
    authLogins.taobao || authLogins.dingtalk || authLogins.douyin || authLogins.feishu ||
    authLogins.twitter
  );
}

function addHttpToUrl(url) {
  if (!url) return '';
  return url.startsWith('http://') || url.startsWith('https://') ? url : `https://${url}`;
}

// 切换密码显示状态
function togglePasswordVisibility(index) {
  if (passwordsList.value && passwordsList.value[index]) {
    // 使用 Vue 的响应式系统更新密码对象
    const updatedPassword = { ...passwordsList.value[index] };
    updatedPassword.showPassword = !updatedPassword.showPassword;
    passwordsList.value[index] = updatedPassword;
  }
}

// 修改主密码
async function changeMasterPassword() {
  try {
    // 清除之前的错误信息
    passwordChangeError.value = '';
    
    // 验证新密码长度
    if (passwordForm.value.newPassword.length < 6) {
      passwordChangeError.value = '新主密码长度至少为6位';
      return;
    }
    
    // 验证两次输入的新密码是否一致
    if (passwordForm.value.newPassword !== passwordForm.value.confirmNewPassword) {
      passwordChangeError.value = '两次输入的新密码不一致';
      return;
    }
    
    // 设置处理中状态
    isChangingPassword.value = true;
    
    // 调用API修改主密码
    const response = await auth.changePassword(
      passwordForm.value.currentPassword, 
      passwordForm.value.newPassword
    );
    
    // 密码修改成功，显示成功消息
    showChangePasswordModal.value = false;
    passwordForm.value = {
      currentPassword: '',
      newPassword: '',
      confirmNewPassword: ''
    };
    
    // 显示成功消息弹窗
    showSuccessModal.value = true;
    successType.value = 'passwordChange';
    
  } catch (error) {
    console.error('修改主密码失败:', error);
    
    // 设置错误信息
    if (error.response && error.response.data && error.response.data.error) {
      passwordChangeError.value = error.response.data.error;
    } else {
      passwordChangeError.value = '修改主密码失败，请稍后再试';
    }
  } finally {
    isChangingPassword.value = false;
  }
}

// 搜索密码
async function searchPasswords() {
  if (!searchQuery.value.trim()) {
    fetchPasswords();
    return;
  }
  
  try {
    const response = await passwords.searchPasswords(searchQuery.value);
    passwordsList.value = response.data;
  } catch (error) {
    console.error('搜索密码失败:', error);
  }
}

// 主密码修改成功后的操作，用户点击确认后执行
function handlePasswordChangeSuccess() {
  showSuccessModal.value = false;
  logout();
}

// 添加监视器，当登录状态变化时确保UI正确更新
watch(isLoggedIn, (newVal) => {
  console.log('isLoggedIn变化:', newVal);
  if (newVal) {
    showLoginForm.value = false; // 如果登录状态为true，确保不显示登录表单
    console.log('已登录，隐藏登录表单');
  } else {
    showLoginForm.value = true; // 如果登录状态为false，确保显示登录表单
    console.log('未登录，显示登录表单');
  }
});

// 监视showLoginForm，以便调试
watch(showLoginForm, (newVal) => {
  console.log('showLoginForm变化:', newVal);
});

// 初始化
onMounted(async () => {
  console.log('PasswordManager组件挂载开始...');
  const token = localStorage.getItem('token');
  
  console.log('组件挂载: token存在?', !!token);
  
  // 初始化passwordsList为空数组而不是undefined
  passwordsList.value = [];
  
  // 首先检查token，再根据结果决定是否需要检查首次设置状态
  if (token) {
    console.log('发现token，验证有效性');
    
    try {
      // 确保Authorization头已设置
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      
      // 验证token有效性 - 使用正确的方法名
      const validationResult = await auth.validate();
      console.log('Token验证结果:', validationResult);
      
      if (validationResult.valid) {
        console.log('Token有效，设置登录状态');
        
        // 更新UI状态
        isLoggedIn.value = true;
        showLoginForm.value = false;
        
        console.log('Token验证成功后状态:', {
          isLoggedIn: isLoggedIn.value,
          showLoginForm: showLoginForm.value
        });
        
        // 使用nextTick确保UI更新
        nextTick(() => {
          console.log('在nextTick中重新检查状态:', {
            isLoggedIn: isLoggedIn.value,
            showLoginForm: showLoginForm.value
          });
          // 获取密码列表
          fetchPasswords();
        });
      } else {
        console.error('Token无效，需要重新登录:', validationResult.error);
        handleInvalidToken();
      }
    } catch (error) {
      console.error('Token验证过程出错:', error);
      handleInvalidToken();
    }
  } else {
    console.log('未发现token，需要登录');
    // 在未找到token时检查是否是首次设置
    await checkFirstTimeSetup();
    showLoginForm.value = true;
    isLoggedIn.value = false;
  }
  
  console.log('PasswordManager组件挂载完成');
});

// 处理无效token的辅助函数
function handleInvalidToken() {
  // 清除token
  localStorage.removeItem('token');
  // 重置认证头
  delete axios.defaults.headers.common['Authorization'];
  
  // 更新UI状态
  showLoginForm.value = true;
  isLoggedIn.value = false;
  // 检查是否首次设置
  checkFirstTimeSetup();
}

// 检查是否首次使用（需要设置主密码）
async function checkFirstTimeSetup() {
  try {
    const response = await auth.checkFirstTimeSetup();
    console.log('首次设置检查响应:', response);
    isFirstTimeSetup.value = response.isFirstTimeSetup;
    console.log('首次设置检查结果:', isFirstTimeSetup.value, '原因:', response.reason);
  } catch (error) {
    console.error('检查首次设置状态失败:', error);
    isFirstTimeSetup.value = false;
  }
}

// 触发文件选择框
function triggerImportFile() {
  fileInput.value.click();
}

// 处理导入的文件
async function handleFileImport(event) {
  const file = event.target.files[0];
  if (!file) return;
  
  const reader = new FileReader();
  
  reader.onload = async (e) => {
    try {
      // 解析CSV文件
      const csv = e.target.result;
      const result = parseCSV(csv);
      
      if (result.length === 0) {
        importError.value = '未检测到有效的密码记录';
        importStep.value = 'error';
        showImportModal.value = true;
        return;
      }
      
      importedPasswords.value = result;
      importStep.value = 'preview';
      showImportModal.value = true;
    } catch (error) {
      console.error('解析CSV文件失败:', error);
      importError.value = '解析CSV文件失败: ' + error.message;
      importStep.value = 'error';
      showImportModal.value = true;
    }
    
    // 重置文件输入，允许选择相同文件
    event.target.value = '';
  };
  
  reader.onerror = (error) => {
    console.error('读取文件失败:', error);
    importError.value = '读取文件失败';
    importStep.value = 'error';
    showImportModal.value = true;
    
    // 重置文件输入
    event.target.value = '';
  };
  
  reader.readAsText(file);
}

// 解析CSV文件
function parseCSV(csv) {
  // 分割为行
  const lines = csv.split('\n');
  if (lines.length <= 1) return [];
  
  // 获取标题行并规范化
  const headers = lines[0].split(',').map(header => 
    header.trim().toLowerCase().replace(/"/g, '')
  );
  
  // 找到对应的列索引
  const nameIndex = headers.indexOf('name');
  const urlIndex = headers.indexOf('url');
  const usernameIndex = headers.indexOf('username');
  const passwordIndex = headers.indexOf('password');
  const noteIndex = headers.indexOf('note');
  const phoneIndex = headers.indexOf('phone');
  const emailIndex = headers.indexOf('email');
  
  // 检查必要的列是否存在
  if (nameIndex === -1 || passwordIndex === -1) {
    throw new Error('CSV文件格式错误：必须包含name和password列');
  }
  
  const parsedPasswords = [];
  
  // 从第二行开始处理数据行
  for (let i = 1; i < lines.length; i++) {
    // 跳过空行
    if (!lines[i].trim()) continue;
    
    // 安全处理CSV中的引号
    let values = [];
    let inQuote = false;
    let currentValue = '';
    let line = lines[i] + ','; // 添加结尾逗号以确保处理最后一个值
    
    for (let j = 0; j < line.length; j++) {
      const char = line[j];
      
      if (char === '"' && (j === 0 || line[j-1] !== '\\')) {
        inQuote = !inQuote;
      } else if (char === ',' && !inQuote) {
        values.push(currentValue.replace(/^"|"$/g, '').replace(/\\"/g, '"'));
        currentValue = '';
      } else {
        currentValue += char;
      }
    }
    
    // 从CSV值映射到导入对象
    const passwordEntry = {
      name: nameIndex >= 0 && nameIndex < values.length ? values[nameIndex].trim() : '',
      website: urlIndex >= 0 && urlIndex < values.length ? values[urlIndex].trim() : '',
      username: usernameIndex >= 0 && usernameIndex < values.length ? values[usernameIndex].trim() : '',
      password: passwordIndex >= 0 && passwordIndex < values.length ? values[passwordIndex].trim() : '',
      notes: noteIndex >= 0 && noteIndex < values.length ? values[noteIndex].trim() : '',
      phone: phoneIndex >= 0 && phoneIndex < values.length ? values[phoneIndex].trim() : '',
      email: emailIndex >= 0 && emailIndex < values.length ? values[emailIndex].trim() : '',
      authLogins: {}
    };
    
    // 只添加有名称的条目
    if (passwordEntry.name) {
      parsedPasswords.push(passwordEntry);
    }
  }
  
  return parsedPasswords;
}

// 关闭导入弹窗
function closeImportModal() {
  showImportModal.value = false;
  importStep.value = 'preview';
  importedPasswords.value = [];
  importProgress.value = 0;
  importError.value = '';
}

// 确认导入密码
async function confirmImport() {
  try {
    importStep.value = 'processing';
    importProgress.value = 0;
    importSuccessCount.value = 0;
    importFailedCount.value = 0;
    
    // 逐个导入密码
    for (let i = 0; i < importedPasswords.value.length; i++) {
      const pwd = importedPasswords.value[i];
      
      try {
        // 调用API保存密码
        await passwords.createPassword({
          name: pwd.name,
          username: pwd.username,
          password: pwd.password,
          website: pwd.website,
          notes: pwd.notes,
          phone: pwd.phone,
          authLogins: {
            google: false,
            wechat: false,
            weibo: false,
            baidu: false,
            facebook: false,
            github: false,
            twitter: false,
            qq: false,
            alipay: false,
            taobao: false,
            dingtalk: false,
            douyin: false,
            feishu: false
          }
        });
        
        importSuccessCount.value++;
      } catch (error) {
        console.error(`导入密码 "${pwd.name}" 失败:`, error);
        importFailedCount.value++;
      }
      
      // 更新进度
      importProgress.value = i + 1;
    }
    
    // 更新界面
    importStep.value = 'complete';
    
    // 刷新密码列表
    await fetchPasswords();
    
  } catch (error) {
    console.error('导入过程中发生错误:', error);
    importError.value = error.message || '导入过程中发生错误';
    importStep.value = 'error';
  }
}

// 重置表单
function resetForm() {
  formData.value = {
    id: null,
    name: '',
    username: '',
    password: '',
    website: '',
    notes: '',
    phone: ''
  };
  resetAuthLogins();
  formErrors.value = {};
}

// 显示成功提示
function showSuccessMessage(msg) {
  successMessage.value = msg;
  successType.value = 'passwordAdd';
  showSuccessModal.value = true;
  setTimeout(() => {
    showSuccessModal.value = false;
  }, 3000);
}

const errorMessage = ref('');
const formErrors = ref({});
const successMessage = ref('');

// Add this among the other ref definitions in the script section
const providerIcons = {
  google: markRaw(defineAsyncComponent(() => import('@/components/icons/GoogleIcon.vue'))),
  wechat: markRaw(defineAsyncComponent(() => import('@/components/icons/WechatIcon.vue'))),
  weibo: markRaw(defineAsyncComponent(() => import('@/components/icons/WeiboIcon.vue'))),
  baidu: markRaw(defineAsyncComponent(() => import('@/components/icons/BaiduIcon.vue'))),
  facebook: markRaw(defineAsyncComponent(() => import('@/components/icons/FacebookIcon.vue'))),
  github: markRaw(defineAsyncComponent(() => import('@/components/icons/GithubIcon.vue'))),
  twitter: markRaw(defineAsyncComponent(() => import('@/components/icons/TwitterIcon.vue'))),
  qq: markRaw(defineAsyncComponent(() => import('@/components/icons/QQIcon.vue'))),
  alipay: markRaw(defineAsyncComponent(() => import('@/components/icons/AlipayIcon.vue'))),
  taobao: markRaw(defineAsyncComponent(() => import('@/components/icons/TaobaoIcon.vue'))),
  dingtalk: markRaw(defineAsyncComponent(() => import('@/components/icons/DingtalkIcon.vue'))),
  douyin: markRaw(defineAsyncComponent(() => import('@/components/icons/DouyinIcon.vue'))),
  feishu: markRaw(defineAsyncComponent(() => import('@/components/icons/FeishuIcon.vue')))
};

// 添加视图模态框处理函数
function openViewModal(password) {
  viewData.value = { ...password };
  // 确保 authLogins 存在
  viewData.value.authLogins = viewData.value.authLogins || {};
  showViewModal.value = true;
  showViewPassword.value = false;
}

// 添加删除确认处理函数
function confirmDeletePassword(id) {
  // 查找要删除的密码对象
  const passwordObj = passwordsList.value.find(p => p.id === id);
  passwordToDelete.value = passwordObj;
  showDeleteModal.value = true;
}

// 添加 mapAuthLoginsToForm 函数
function mapAuthLoginsToForm(authLogins = {}) {
  // 重置所有登录状态
  resetAuthLogins();
  
  // 设置已启用的登录项
  Object.keys(authLogins).forEach(provider => {
    if (authLogins[provider] && selectedAuthLogins.value.hasOwnProperty(provider)) {
      selectedAuthLogins.value[provider] = true;
    }
  });
}

// 重置授权登录选项
function resetAuthLogins() {
  Object.keys(selectedAuthLogins.value).forEach(key => {
    selectedAuthLogins.value[key] = false;
  });
}

// 在 ref 声明区域添加
const showRestoreBackupModal = ref(false);
const hasBackupData = ref(false);
const backupPasswordsCount = ref(0);

// 添加检查备份数据函数
function checkBackupData() {
  try {
    const backupJson = localStorage.getItem('password_backup');
    if (backupJson) {
      const backupData = JSON.parse(backupJson);
      if (Array.isArray(backupData) && backupData.length > 0) {
        hasBackupData.value = true;
        backupPasswordsCount.value = backupData.length;
        
        // 如果备份数据明显多于当前数据，提示用户恢复
        if (passwordsList.value.length < backupData.length * 0.7) {
          const lostCount = backupData.length - passwordsList.value.length;
          const shouldRestore = confirm(
            `检测到您可能丢失了约 ${lostCount} 条密码数据。\n\n` +
            `系统找到了一个包含 ${backupData.length} 条记录的备份，是否恢复？`
          );
          
          if (shouldRestore) {
            restoreBackupData();
          }
        }
      }
    }
  } catch (error) {
    console.error('检查备份数据失败:', error);
  }
}

// 添加恢复备份数据函数
async function restoreBackupData() {
  try {
    const backupJson = localStorage.getItem('password_backup');
    if (!backupJson) {
      message.error('未找到有效的备份数据');
      return;
    }
    
    const backupData = JSON.parse(backupJson);
    if (!Array.isArray(backupData) || backupData.length === 0) {
      message.error('备份数据格式无效或为空');
      return;
    }
    
    // 显示确认对话框
    if (!confirm(`您确定要恢复 ${backupData.length} 条密码数据吗？此操作将覆盖当前数据！`)) {
      return;
    }
    
    // 执行恢复操作
    let success = 0;
    let failed = 0;
    const total = backupData.length;
    
    message.info(`开始恢复 ${total} 条密码数据...`);
    
    // 先删除当前所有密码
    for (const pwd of passwordsList.value) {
      try {
        await passwords.deletePassword(pwd.id);
      } catch (error) {
        console.error(`删除现有密码 ${pwd.id} 失败:`, error);
      }
    }
    
    // 逐个创建备份的密码
    for (let i = 0; i < backupData.length; i++) {
      const pwd = backupData[i];
      try {
        // 创建新密码
        await passwords.createPassword({
          name: pwd.name,
          username: pwd.username,
          password: pwd.password,
          website: pwd.website, 
          notes: pwd.notes,
          phone: pwd.phone,
          authLogins: pwd.authLogins || {}
        });
        success++;
      } catch (error) {
        console.error(`恢复密码 "${pwd.name}" 失败:`, error);
        failed++;
      }
      
      // 每恢复10条显示一次进度
      if (i % 10 === 0 || i === backupData.length - 1) {
        message.info(`正在恢复: ${i + 1}/${total}`);
      }
    }
    
    // 完成恢复
    message.success(`恢复完成！成功: ${success}，失败: ${failed}`);
    
    // 刷新密码列表
    await fetchPasswords();
    
    // 关闭恢复模态窗口
    showRestoreBackupModal.value = false;
    
    // 如果全部恢复成功，可以清除备份
    if (success === total) {
      if (confirm('备份数据已成功恢复，是否删除备份？')) {
        localStorage.removeItem('password_backup');
        hasBackupData.value = false;
      }
    }
  } catch (error) {
    console.error('恢复备份数据失败:', error);
    message.error('恢复备份数据失败: ' + error.message);
  }
}

// 认证登录图标编辑模态窗口
const showAuthLoginsModal = ref(false);
const editingAuthLogins = ref({
  google: false,
  wechat: false,
  weibo: false,
  baidu: false,
  facebook: false,
  github: false,
  qq: false,
  alipay: false,
  taobao: false,
  dingtalk: false,
  douyin: false,
  feishu: false,
  twitter: false
});

// 添加保存认证登录的函数
function saveAuthLogins() {
  // 将编辑后的状态保存到 selectedAuthLogins
  selectedAuthLogins.value = { ...editingAuthLogins.value };
  showAuthLoginsModal.value = false;
}

// 添加 authProviders 数据
const authProviders = [
  { name: '谷歌', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/GoogleIcon.vue'))) },
  { name: '微信', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/WechatIcon.vue'))) },
  { name: '微博', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/WeiboIcon.vue'))) },
  { name: '百度', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/BaiduIcon.vue'))) },
  { name: 'Facebook', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/FacebookIcon.vue'))) },
  { name: 'Github', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/GithubIcon.vue'))) },
  { name: 'X(Twitter)', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/TwitterIcon.vue'))) },
  { name: 'QQ', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/QQIcon.vue'))) },
  { name: '支付宝', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/AlipayIcon.vue'))) },
  { name: '淘宝', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/TaobaoIcon.vue'))) },
  { name: '钉钉', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/DingtalkIcon.vue'))) },
  { name: '抖音', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/DouyinIcon.vue'))) },
  { name: '飞书', icon: markRaw(defineAsyncComponent(() => import('@/components/icons/FeishuIcon.vue'))) }
];

</script>
