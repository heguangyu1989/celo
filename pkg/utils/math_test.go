package utils

import (
	"testing"
)

func TestMaxInt(t *testing.T) {
	// 测试用例1: 第一个数大于第二个数
	result := MaxInt(5, 3)
	if result != 5 {
		t.Errorf("MaxInt(5, 3) = %d; 期望 5", result)
	}
	
	// 测试用例2: 第二个数大于第一个数
	result = MaxInt(2, 8)
	if result != 8 {
		t.Errorf("MaxInt(2, 8) = %d; 期望 8", result)
	}
	
	// 测试用例3: 两个数相等
	result = MaxInt(4, 4)
	if result != 4 {
		t.Errorf("MaxInt(4, 4) = %d; 期望 4", result)
	}
	
	// 测试用例4: 负数比较，第一个数大于第二个数
	result = MaxInt(-1, -5)
	if result != -1 {
		t.Errorf("MaxInt(-1, -5) = %d; 期望 -1", result)
	}
	
	// 测试用例5: 负数比较，第二个数大于第一个数
	result = MaxInt(-10, -2)
	if result != -2 {
		t.Errorf("MaxInt(-10, -2) = %d; 期望 -2", result)
	}
	
	// 测试用例6: 正数和负数比较
	result = MaxInt(-5, 5)
	if result != 5 {
		t.Errorf("MaxInt(-5, 5) = %d; 期望 5", result)
	}
	
	// 测试用例7: 零和正数比较
	result = MaxInt(0, 10)
	if result != 10 {
		t.Errorf("MaxInt(0, 10) = %d; 期望 10", result)
	}
	
	// 测试用例8: 零和负数比较
	result = MaxInt(0, -10)
	if result != 0 {
		t.Errorf("MaxInt(0, -10) = %d; 期望 0", result)
	}
}