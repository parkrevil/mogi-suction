package packet

// parseAttack - 공격 관련 패킷 파싱
func parseAttack(data []byte) interface{} {
	println("parseAttack: &x", data)
	return nil
}

// parseAction - 액션 관련 패킷 파싱
func parseAction(data []byte) interface{} {
	println("parseAction: &x", data)
	return nil
}

// parseDamage - 데미지 관련 패킷 파싱
func parseDamage(data []byte) interface{} {
	println("parseDamage: &x", data)
	return nil
}

// parseHPChanged - HP 변경 관련 패킷 파싱
func parseHPChanged(data []byte) interface{} {
	println("parseHPChanged: &x", data)
	return nil
}

// parseSelfDamage - 자기 데미지 관련 패킷 파싱
func parseSelfDamage(data []byte) interface{} {
	println("parseSelfDamage: &x", data)
	return nil
}

// parseItem - 아이템 관련 패킷 파싱
func parseItem(data []byte) interface{} {
	println("parseItem: &x", data)
	return nil
}
