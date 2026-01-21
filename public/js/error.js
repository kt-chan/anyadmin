// 错误页面特定脚本
document.addEventListener('DOMContentLoaded', function() {
  if (window.pageError) {
    console.error('页面错误:', window.pageError);
  }
});
