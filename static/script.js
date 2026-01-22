    const navBtn = document.getElementById('navSearchBtn');
    const heroInput = document.getElementById('heroSearchInput');

    navBtn.addEventListener('click', function() {
        // Menggunakan window.scrollTo dengan opsi smooth agar lebih stabil di semua browser
        const targetPosition = heroInput.getBoundingClientRect().top + window.pageYOffset - 100; // -100 agar tidak mepet navbar fixed
        
        window.scrollTo({
            top: targetPosition,
            behavior: 'smooth' // Ini yang bikin efek meluncur/scroll bray
        });
        
        // Fokuskan kursor setelah animasi scroll kira-kira selesai
        setTimeout(() => {
            heroInput.focus();
        }, 800); 
    });
