package com.storytrim.app.ui.home

import android.os.Bundle
import android.view.View
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import com.storytrim.app.R
import com.storytrim.app.databinding.ActivityHomeBinding
import com.storytrim.app.ui.home.tab.ProfileFragment
import com.storytrim.app.ui.home.tab.ShelfFragment
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class HomeActivity : AppCompatActivity() {

    private lateinit var binding: ActivityHomeBinding
    private val shelfFragment = ShelfFragment()
    private val profileFragment = ProfileFragment()
    private var activeTab: Tab = Tab.SHELF

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        WindowCompat.setDecorFitsSystemWindows(window, false)
        binding = ActivityHomeBinding.inflate(layoutInflater)
        setContentView(binding.root)

        setupInsets()
        setupTabs()
        if (savedInstanceState == null) {
            supportFragmentManager.beginTransaction()
                .add(R.id.contentContainer, shelfFragment, Tab.SHELF.tag)
                .commit()
        }
        updateTabUi(Tab.SHELF)
    }

    private fun setupInsets() {
        ViewCompat.setOnApplyWindowInsetsListener(binding.root) { view, windowInsets ->
            val insets = windowInsets.getInsets(WindowInsetsCompat.Type.systemBars())
            binding.bottomNav.setPadding(
                binding.bottomNav.paddingLeft,
                binding.bottomNav.paddingTop,
                binding.bottomNav.paddingRight,
                insets.bottom
            )
            view.setPadding(0, 0, 0, 0)
            WindowInsetsCompat.CONSUMED
        }
    }

    private fun setupTabs() {
        binding.tabShelf.setOnClickListener { switchTab(Tab.SHELF) }
        binding.tabProfile.setOnClickListener { switchTab(Tab.PROFILE) }
    }

    private fun switchTab(target: Tab) {
        if (target == activeTab) return
        activeTab = target

        val fragment = when (target) {
            Tab.SHELF -> shelfFragment
            Tab.PROFILE -> profileFragment
        }

        val transaction = supportFragmentManager.beginTransaction()
        supportFragmentManager.fragments.forEach { transaction.hide(it) }
        if (supportFragmentManager.findFragmentByTag(target.tag) == null) {
            transaction.add(R.id.contentContainer, fragment, target.tag)
        } else {
            transaction.show(fragment)
        }
        transaction.commit()
        updateTabUi(target)
    }

    private fun updateTabUi(tab: Tab) {
        val isShelf = tab == Tab.SHELF
        binding.tabShelf.isSelected = isShelf
        binding.tabProfile.isSelected = !isShelf

        val activeColor = getColor(com.storytrim.app.R.color.storytrim_text_primary)
        val inactiveColor = getColor(com.storytrim.app.R.color.storytrim_text_secondary)
        binding.tabShelfLabel.setTextColor(if (isShelf) activeColor else inactiveColor)
        binding.tabProfileLabel.setTextColor(if (isShelf) inactiveColor else activeColor)
        binding.tabShelfIcon.setColorFilter(if (isShelf) activeColor else inactiveColor)
        binding.tabProfileIcon.setColorFilter(if (isShelf) inactiveColor else activeColor)
        binding.tabShelfIndicator.visibility = if (isShelf) View.VISIBLE else View.INVISIBLE
        binding.tabProfileIndicator.visibility = if (isShelf) View.INVISIBLE else View.VISIBLE

        animateTab(binding.tabShelfIcon, binding.tabShelfLabel, isShelf)
        animateTab(binding.tabProfileIcon, binding.tabProfileLabel, !isShelf)
    }

    private fun animateTab(icon: View, label: View, active: Boolean) {
        val scale = if (active) 1.05f else 1f
        icon.animate().scaleX(scale).scaleY(scale).setDuration(180).start()
        label.animate().alpha(if (active) 1f else 0.7f).setDuration(180).start()
    }

    enum class Tab(val tag: String) {
        SHELF("tab_shelf"),
        PROFILE("tab_profile")
    }
}
