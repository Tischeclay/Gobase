// passwordgen.go
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
)

type PasswordGenerator struct {
	length     int
	count      int
	useUpper   bool
	useLower   bool
	useDigits  bool
	useSymbols bool
	exclude    string
	minUpper   int
	minLower   int
	minDigits  int
	minSymbols int
}

func NewPasswordGenerator() *PasswordGenerator {
	return &PasswordGenerator{
		length:     16,
		count:      1,
		useUpper:   true,
		useLower:   true,
		useDigits:  true,
		useSymbols: false,
		minUpper:   0,
		minLower:   0,
		minDigits:  0,
		minSymbols: 0,
	}
}

func (pg *PasswordGenerator) generate() (string, error) {
	// 构建字符集
	var chars []rune

	if pg.useUpper {
		chars = append(chars, []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")...)
	}
	if pg.useLower {
		chars = append(chars, []rune("abcdefghijklmnopqrstuvwxyz")...)
	}
	if pg.useDigits {
		chars = append(chars, []rune("0123456789")...)
	}
	if pg.useSymbols {
		chars = append(chars, []rune("!@#$%^&*()-_=+[]{}|;:,.<>?")...)
	}

	if len(chars) == 0 {
		return "", fmt.Errorf("至少需要选择一种字符类型")
	}

	// 移除排除的字符
	if pg.exclude != "" {
		filtered := make([]rune, 0)
		excludeSet := make(map[rune]bool)
		for _, c := range pg.exclude {
			excludeSet[c] = true
		}

		for _, c := range chars {
			if !excludeSet[c] {
				filtered = append(filtered, c)
			}
		}
		chars = filtered
	}

	if len(chars) == 0 {
		return "", fmt.Errorf("排除后没有可用字符")
	}

	maxAttempts := 100
	for attempt := 0; attempt < maxAttempts; attempt++ {
		password, err := pg.tryGenerate(chars)
		if err == nil {
			return password, nil
		}
	}

	return "", fmt.Errorf("无法生成满足要求的密码")
}

func (pg *PasswordGenerator) tryGenerate(chars []rune) (string, error) {
	password := make([]rune, pg.length)

	// 统计各类字符数量
	counts := map[string]int{
		"upper":  0,
		"lower":  0,
		"digit":  0,
		"symbol": 0,
	}

	for i := 0; i < pg.length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}

		c := chars[n.Int64()]
		password[i] = c

		// 计数
		if c >= 'A' && c <= 'Z' {
			counts["upper"]++
		} else if c >= 'a' && c <= 'z' {
			counts["lower"]++
		} else if c >= '0' && c <= '9' {
			counts["digit"]++
		} else {
			counts["symbol"]++
		}
	}

	// 检查最小要求
	if counts["upper"] < pg.minUpper ||
		counts["lower"] < pg.minLower ||
		counts["digit"] < pg.minDigits ||
		counts["symbol"] < pg.minSymbols {
		return "", fmt.Errorf("不满足最小要求")
	}

	return string(password), nil
}

func (pg *PasswordGenerator) generateMultiple() ([]string, error) {
	passwords := make([]string, pg.count)
	seen := make(map[string]bool)

	for i := 0; i < pg.count; i++ {
		maxAttempts := 10
		generated := false

		for attempt := 0; attempt < maxAttempts; attempt++ {
			pwd, err := pg.generate()
			if err != nil {
				continue
			}

			if !seen[pwd] {
				passwords[i] = pwd
				seen[pwd] = true
				generated = true
				break
			}
		}

		if !generated {
			return nil, fmt.Errorf("无法生成足够的唯一密码")
		}
	}

	return passwords, nil
}

func (pg *PasswordGenerator) estimateEntropy() float64 {
	// 计算熵值
	var poolSize int

	if pg.useUpper {
		poolSize += 26
	}
	if pg.useLower {
		poolSize += 26
	}
	if pg.useDigits {
		poolSize += 10
	}
	if pg.useSymbols {
		poolSize += 32
	}

	if poolSize == 0 {
		return 0
	}

	return float64(pg.length) * (float64(poolSize) / float64(26))
}

func main() {
	pg := NewPasswordGenerator()

	// 定义参数
	length := flag.Int("l", 16, "密码长度")
	count := flag.Int("n", 1, "生成数量")
	noUpper := flag.Bool("no-upper", false, "不使用大写字母")
	noLower := flag.Bool("no-lower", false, "不使用小写字母")
	noDigits := flag.Bool("no-digits", false, "不使用数字")
	useSymbols := flag.Bool("s", false, "使用特殊字符")
	exclude := flag.String("e", "", "排除的字符")
	minUpper := flag.Int("min-upper", 0, "最少大写字母数")
	minLower := flag.Int("min-lower", 0, "最少小写字母数")
	minDigits := flag.Int("min-digits", 0, "最少数字数")
	minSymbols := flag.Int("min-symbols", 0, "最少特殊字符数")

	flag.Parse()

	pg.length = *length
	pg.count = *count
	pg.useUpper = !*noUpper
	pg.useLower = !*noLower
	pg.useDigits = !*noDigits
	pg.useSymbols = *useSymbols
	pg.exclude = *exclude
	pg.minUpper = *minUpper
	pg.minLower = *minLower
	pg.minDigits = *minDigits
	pg.minSymbols = *minSymbols

	// 验证最小要求
	if pg.minUpper+pg.minLower+pg.minDigits+pg.minSymbols > pg.length {
		fmt.Fprintf(os.Stderr, "错误: 最小要求总和超过密码长度\n")
		os.Exit(1)
	}

	// 生成密码
	passwords, err := pg.generateMultiple()
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成失败: %v\n", err)
		os.Exit(1)
	}

	// 计算熵值
	entropy := pg.estimateEntropy()

	// 输出
	fmt.Println("生成的密码:")
	//fmt.Println("=" * 40)

	for i, pwd := range passwords {
		fmt.Printf("%2d. %s\n", i+1, pwd)
	}
	
}
