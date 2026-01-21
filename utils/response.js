const success = (res, data = {}, message = 'Success', statusCode = 200) => {
  return res.status(statusCode).json({
    success: true,
    message,
    data
  });
};

const error = (res, message = 'Internal Server Error', statusCode = 500, errorDetails = null) => {
  const response = {
    success: false,
    message
  };
  
  if (errorDetails && process.env.NODE_ENV === 'development') {
    response.error = errorDetails;
  }
  
  return res.status(statusCode).json(response);
};

module.exports = {
  success,
  error
};
