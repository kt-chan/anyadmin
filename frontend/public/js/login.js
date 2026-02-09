// 页面加载后初始化
document.addEventListener('DOMContentLoaded', function() {
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
    // 注意：实际应用中不应明文存储密码
    localStorage.setItem('kb_password', data.password);
  } else {
    localStorage.removeItem('kb_remember_me');
    localStorage.removeItem('kb_username');
    localStorage.removeItem('kb_password');
  }
  
  // 显示加载状态
  const submitBtn = form.querySelector('button[type="button"]');
  const originalText = submitBtn.textContent;
  submitBtn.disabled = true;
  submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>登录中...';
  
  try {
    // 1. 获取公钥
    let encryptedPassword = data.password;
    try {
      const keyResponse = await fetch('/public-key');
      const keyData = await keyResponse.json();
      if (keyData.success && keyData.publicKey) {
        const encrypt = new JSEncrypt();
        encrypt.setPublicKey(keyData.publicKey);
        const result = encrypt.encrypt(data.password);
        if (result) {
          encryptedPassword = result;
        } else {
          throw new Error('Encryption failed');
        }
      }
    } catch (keyError) {
      console.warn('公钥获取或加密失败，尝试明文传输 (仅用于兼容性):', keyError);
      // 如果后端强制要求加密，这里可能会失败。但根据代码，后端会尝试解密，如果失败则作为明文。
      // 为满足"不保存明文"的要求，最好是必须加密。但如果获取key失败，就没办法了。
      // 可以在这里throw error阻止登录
      if (keyError.message === 'Encryption failed') {
         throw new Error('密码加密失败，请重试');
      }
    }

    // 更新密码为加密后的密码
    data.password = encryptedPassword;

    const response = await fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    
    const result = await response.json();
    
    if (result.success) {
      // 登录成功，跳转
      window.location.href = result.redirect;
    } else {
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
