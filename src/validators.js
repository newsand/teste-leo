function isValidEmail(email) {
  if (typeof email !== 'string') return false;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

function validateCreateUser(body) {
  const errors = [];

  if (!body || typeof body !== 'object') {
    return { valid: false, errors: ['Request body must be a JSON object'] };
  }

  if (!body.name || typeof body.name !== 'string' || body.name.trim() === '') {
    errors.push('name is required and must be a non-empty string');
  }

  if (!body.email || typeof body.email !== 'string') {
    errors.push('email is required and must be a string');
  } else if (!isValidEmail(body.email)) {
    errors.push('email must be a valid email address');
  }

  if (!body.password || typeof body.password !== 'string' || body.password.length < 1) {
    errors.push('password is required and must be a non-empty string');
  }

  return {
    valid: errors.length === 0,
    errors
  };
}

function validateUpdateUser(body) {
  const errors = [];

  if (!body || typeof body !== 'object') {
    return { valid: false, errors: ['Request body must be a JSON object'] };
  }

  const hasAtLeastOneField = body.name !== undefined || body.email !== undefined || body.password !== undefined;
  if (!hasAtLeastOneField) {
    errors.push('At least one field (name, email, or password) must be provided');
  }

  if (body.name !== undefined && (typeof body.name !== 'string' || body.name.trim() === '')) {
    errors.push('name must be a non-empty string');
  }

  if (body.email !== undefined) {
    if (typeof body.email !== 'string') {
      errors.push('email must be a string');
    } else if (!isValidEmail(body.email)) {
      errors.push('email must be a valid email address');
    }
  }

  if (body.password !== undefined && (typeof body.password !== 'string' || body.password.length < 1)) {
    errors.push('password must be a non-empty string');
  }

  return {
    valid: errors.length === 0,
    errors
  };
}

module.exports = { isValidEmail, validateCreateUser, validateUpdateUser };
