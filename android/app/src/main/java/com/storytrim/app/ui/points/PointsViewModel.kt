package com.storytrim.app.ui.points

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.storytrim.app.data.dto.PointsLedgerItem
import com.storytrim.app.data.repository.PointsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PointsViewModel @Inject constructor(
    private val pointsRepository: PointsRepository
) : ViewModel() {

    private val _balance = MutableLiveData(0)
    val balance: LiveData<Int> = _balance

    private val _records = MutableLiveData<List<PointsLedgerItem>>(emptyList())
    val records: LiveData<List<PointsLedgerItem>> = _records

    private val _isLoading = MutableLiveData(false)
    val isLoading: LiveData<Boolean> = _isLoading

    private val _error = MutableLiveData<String?>(null)
    val error: LiveData<String?> = _error

    fun load() {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            val balanceResult = pointsRepository.getBalance()
            if (balanceResult.isSuccess) {
                _balance.value = balanceResult.getOrNull()?.balance ?: 0
            } else {
                _error.value = balanceResult.exceptionOrNull()?.message
            }

            val ledgerResult = pointsRepository.getLedger(1, 30)
            if (ledgerResult.isSuccess) {
                _records.value = ledgerResult.getOrNull()?.items ?: emptyList()
            } else {
                _error.value = ledgerResult.exceptionOrNull()?.message
            }
            _isLoading.value = false
        }
    }
}
