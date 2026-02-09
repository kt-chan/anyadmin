// 登录处理函数
async function handleLogin() {
  const form = document.getElementById('login-form');
  const formData = new FormData(form);
  const data = Object.fromEntries(formData.entries());
  const rememberMe = formData.get('remember') === 'on';
  
  // 记住我功能
  if (rememberMe) {
    localStorage.setItem('kb_remember_me', 'true');
    localStorage.setItem('kb_username', data.username);
    localStorage.setItem('kb_password', data.password);
  } else {
    localStorage.removeItem('kb_remember_me');
    localStorage.removeItem('kb_username');
    localStorage.removeItem('kb_password');
  }
  
  // 显示加载状态
  const submitBtn = form.querySelector('button');
  const originalText = submitBtn.textContent;
  submitBtn.disabled = true;
  submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>登录中...';
  
  try {
    console.log('开始登录流程...');
    // 1. 获取公钥
    let encryptedPassword = data.password;
    try {
      const keyResponse = await fetch('/public-key');
      const keyData = await keyResponse.json();
      if (keyData.success && keyData.publicKey) {
        console.log('获取公钥成功，正在加密...');
        const encrypt = new JSEncrypt();
        encrypt.setPublicKey(keyData.publicKey);
        const result = encrypt.encrypt(data.password);
        if (result) {
          encryptedPassword = result;
          console.log('加密成功');
        } else {
          throw new Error('Encryption failed');
        }
      }
    } catch (keyError) {
      console.warn('公钥获取或加密失败:', keyError);
      throw new Error('密码加密失败，请重试');
    }

    // 更新密码为加密后的密码
    data.password = encryptedPassword;

    console.log('发送登录请求到服务器...');
    const response = await fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    
    const result = await response.json();
    
    if (result.success) {
      console.log('登录成功，准备跳转');
      // 保存登录状态
      localStorage.setItem('kb_user_logged_in', 'true');
      window.location.href = result.redirect;
    } else {
      console.warn('登录失败:', result.message);
      alert(result.message || '登录失败');
      submitBtn.disabled = false;
      submitBtn.textContent = originalText;
    }
  } catch (error) {
    console.error('登录错误:', error);
    alert(error.message || '网络错误，请重试');
    submitBtn.disabled = false;
    submitBtn.textContent = originalText;
  }
}

// Ensure it is globally available
window.handleLogin = handleLogin;

// 页面加载后初始化
document.addEventListener('DOMContentLoaded', function() {
  console.log('Login JS Loaded');
  const loginForm = document.getElementById('login-form');
  if (loginForm) {
    loginForm.addEventListener('submit', function(e) {
      e.preventDefault();
      handleLogin();
    });
    
    // 回车键支持
    const inputs = loginForm.querySelectorAll('input');
    inputs.forEach(input => {
      input.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
          e.preventDefault();
          handleLogin();
        }
      });
    });
  }
  
  // 记住我功能
  const rememberMe = localStorage.getItem('kb_remember_me');
  if (rememberMe === 'true') {
    const usernameInput = document.querySelector('input[name="username"]');
    const passwordInput = document.querySelector('input[name="password"]');
    const rememberCheckbox = document.querySelector('input[name="remember"]');
    
    if (usernameInput) usernameInput.value = localStorage.getItem('kb_username') || '';
    if (passwordInput) passwordInput.value = localStorage.getItem('kb_password') || '';
    if (rememberCheckbox) rememberCheckbox.checked = true;
  }
});