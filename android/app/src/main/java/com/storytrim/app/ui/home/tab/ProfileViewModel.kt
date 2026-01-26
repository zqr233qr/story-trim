package com.storytrim.app.ui.home.tab

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.asLiveData
import androidx.lifecycle.viewModelScope
import com.storytrim.app.data.repository.AuthRepository
import com.storytrim.app.data.repository.PointsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ProfileViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val pointsRepository: PointsRepository
) : ViewModel() {

    val username: LiveData<String> = authRepository.usernameFlow()
        .map { it.ifBlank { "" } }
        .asLiveData()

    private val _logoutResult = MutableLiveData<Boolean>()
    val logoutResult: LiveData<Boolean> = _logoutResult

    private val _pointsBalance = MutableLiveData(0)
    val pointsBalance: LiveData<Int> = _pointsBalance

    fun logout() {
        viewModelScope.launch {
            authRepository.logout()
            _logoutResult.value = true
        }
    }

    fun loadPoints(isLoggedIn: Boolean) {
        viewModelScope.launch {
            if (!isLoggedIn) {
                _pointsBalance.value = 0
                return@launch
            }
            val result = pointsRepository.getBalance()
            if (result.isSuccess) {
                _pointsBalance.value = result.getOrNull()?.balance ?: 0
            }
        }
    }
}
