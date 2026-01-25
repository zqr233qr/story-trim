package com.storytrim.app.ui.login

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.storytrim.app.data.model.User
import com.storytrim.app.data.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _loginResult = MutableLiveData<Result<User>>()
    val loginResult: LiveData<Result<User>> = _loginResult

    private val _registerResult = MutableLiveData<Result<Unit>>()
    val registerResult: LiveData<Result<Unit>> = _registerResult

    private val _isLoading = MutableLiveData<Boolean>()
    val isLoading: LiveData<Boolean> = _isLoading

    fun login(username: String, password: String) {
        _isLoading.value = true
        viewModelScope.launch {
            val result = authRepository.login(username, password)
            _loginResult.value = result
            _isLoading.value = false
        }
    }

    fun register(username: String, password: String) {
        _isLoading.value = true
        viewModelScope.launch {
            val result = authRepository.register(username, password)
            _registerResult.value = result
            _isLoading.value = false
        }
    }
}
