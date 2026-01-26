package com.storytrim.app.ui.reader.dialog

import android.content.Intent
import android.os.Build
import android.os.Bundle
import android.os.CountDownTimer
import android.provider.Settings
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.SeekBar
import com.storytrim.app.ui.common.ToastHelper
import androidx.fragment.app.activityViewModels
import com.google.android.material.bottomsheet.BottomSheetDialogFragment
import com.storytrim.app.R
import com.storytrim.app.databinding.DialogTtsControlPanelBinding
import com.storytrim.app.ui.reader.ReaderViewModel
import com.storytrim.app.ui.reader.tts.FloatingLyricsManager
import com.storytrim.app.ui.reader.tts.TtsController

class TtsControlPanelDialog : BottomSheetDialogFragment() {

    private var _binding: DialogTtsControlPanelBinding? = null
    private val binding get() = _binding!!

    private val viewModel: ReaderViewModel by activityViewModels()

    private var currentSpeechRate = 1.0f
    private var floatingLyricsManager: FloatingLyricsManager? = null
    private var timer: CountDownTimer? = null

    private val prefs by lazy {
        requireContext().getSharedPreferences("tts_prefs", android.content.Context.MODE_PRIVATE)
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = DialogTtsControlPanelBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        // 初始化悬浮窗管理器
        viewModel.ttsController.value?.let { controller ->
            floatingLyricsManager = controller.floatingLyricsManager
        }

        restorePreferences()
        setupClickListeners()
        observeTtsState()
    }

    private fun restorePreferences() {
        val savedRate = prefs.getFloat("tts_speech_rate", 1.0f).coerceIn(0.5f, 3.0f)
        currentSpeechRate = savedRate
        val progress = ((savedRate - 0.5f) / 2.5f * 25f).toInt().coerceIn(0, 25)
        binding.seekBarSpeed.progress = progress
        binding.tvSpeedValue.text = String.format("%.1fx", savedRate)
        viewModel.ttsController.value?.setSpeechRate(savedRate)

        val minutes = prefs.getInt("tts_timer_minutes", 0).coerceAtLeast(0)
        binding.seekBarTimer.progress = (minutes / 5).coerceIn(0, 25)
        if (minutes > 0) {
            startTimer(minutes)
        } else {
            binding.tvTimerValue.text = "关闭"
        }
    }

    private fun setupClickListeners() {
        binding.btnPlayPause.setOnClickListener {
            viewModel.ttsController.value?.togglePlayPause()
        }

        binding.btnPrevious.setOnClickListener {
            viewModel.ttsController.value?.previous()
        }

        binding.btnNext.setOnClickListener {
            viewModel.ttsController.value?.next()
        }

        // 桌面歌词按钮点击事件
        binding.btnFloatingLyrics.setOnClickListener {
            handleFloatingLyricsClick()
        }

        binding.seekBarSpeed.setOnSeekBarChangeListener(object : SeekBar.OnSeekBarChangeListener {
            override fun onProgressChanged(seekBar: SeekBar?, progress: Int, fromUser: Boolean) {
                if (fromUser) {
                    val rate = 0.5f + (progress * 2.5f / 25f)
                    currentSpeechRate = rate
                    binding.tvSpeedValue.text = String.format("%.1fx", rate)
                    viewModel.ttsController.value?.setSpeechRate(rate)
                    prefs.edit().putFloat("tts_speech_rate", rate).apply()
                }
            }

            override fun onStartTrackingTouch(seekBar: SeekBar?) {}

            override fun onStopTrackingTouch(seekBar: SeekBar?) {}
        })

        binding.seekBarTimer.setOnSeekBarChangeListener(object : SeekBar.OnSeekBarChangeListener {
            override fun onProgressChanged(seekBar: SeekBar?, progress: Int, fromUser: Boolean) {
                binding.tvTimerValue.text = formatTimerValue(progress)
            }

            override fun onStartTrackingTouch(seekBar: SeekBar?) {}

            override fun onStopTrackingTouch(seekBar: SeekBar?) {
                val minutes = (seekBar?.progress ?: 0) * 5
                prefs.edit().putInt("tts_timer_minutes", minutes).apply()
                startTimer(minutes)
            }
        })
    }

    /**
     * 处理桌面歌词按钮点击
     */
    private fun handleFloatingLyricsClick() {
        val manager = floatingLyricsManager ?: return

        if (manager.canDrawOverlays()) {
            // 有权限，直接显示悬浮窗
            manager.show()
            ToastHelper.show(requireContext(), "桌面歌词已开启")
        } else {
            // 没有权限，请求权限
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                val intent = Intent(Settings.ACTION_MANAGE_OVERLAY_PERMISSION)
                startActivity(intent)
                ToastHelper.show(
                    requireContext(),
                    "请开启悬浮窗权限后重试",
                    long = true
                )
            }
        }
    }

    private fun observeTtsState() {
        viewModel.ttsController.value?.ttsState?.observe(viewLifecycleOwner) { state ->
            when (state) {
                TtsController.TtsState.PLAYING -> updatePlayPauseIcon(true)
                TtsController.TtsState.PAUSED -> updatePlayPauseIcon(false)
                TtsController.TtsState.STOPPED -> updatePlayPauseIcon(false)
                TtsController.TtsState.IDLE -> updatePlayPauseIcon(false)
                TtsController.TtsState.ERROR -> updatePlayPauseIcon(false)
                null -> {}
            }
        }
    }

    private fun formatTimerValue(progress: Int): String {
        if (progress <= 0) return "关闭"
        return "${progress * 5}分钟"
    }

    private fun startTimer(minutes: Int) {
        timer?.cancel()
        if (minutes <= 0) {
            binding.tvTimerValue.text = "关闭"
            prefs.edit().remove("tts_timer_minutes").apply()
            return
        }
        val totalMs = minutes * 60_000L
        timer = object : CountDownTimer(totalMs, 60_000L) {
            override fun onTick(millisUntilFinished: Long) {
                val remainMinutes = (millisUntilFinished / 60_000L).toInt().coerceAtLeast(1)
                binding.tvTimerValue.text = "${remainMinutes}分钟"
            }

            override fun onFinish() {
                binding.tvTimerValue.text = "关闭"
                viewModel.ttsController.value?.stop()
                binding.seekBarTimer.progress = 0
            }
        }.start()
        binding.tvTimerValue.text = "${minutes}分钟"
    }

    private fun updatePlayPauseIcon(isPlaying: Boolean) {
        binding.btnPlayPause.setImageResource(
            if (isPlaying) R.drawable.ic_pause else R.drawable.ic_play
        )
    }

    override fun onDestroyView() {
        super.onDestroyView()
        timer?.cancel()
        timer = null
        _binding = null
    }

    companion object {
        const val TAG = "TtsControlPanelDialog"

        fun newInstance(): TtsControlPanelDialog {
            return TtsControlPanelDialog()
        }
    }
}
