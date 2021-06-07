package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
)

func main() {
	ips := []net.IP{
		// 21
		net.ParseIP("176.31.205.144"),
		net.ParseIP("176.31.205.146"),
		net.ParseIP("176.31.205.147"),
		net.ParseIP("176.31.205.148"),
		net.ParseIP("176.31.205.149"),
		net.ParseIP("176.31.205.150"),
		net.ParseIP("176.31.205.151"),
		net.ParseIP("162.221.206.2"),
		net.ParseIP("162.221.206.4"),
		net.ParseIP("162.221.206.6"),
		net.ParseIP("162.221.206.7"),
		net.ParseIP("94.23.77.74"),
		net.ParseIP("162.221.206.5"),
		net.ParseIP("162.221.206.8"),
		net.ParseIP("162.221.206.3"),
		net.ParseIP("94.23.77.73"),
		net.ParseIP("176.31.205.152"),
		net.ParseIP("176.31.205.153"),
		net.ParseIP("176.31.205.155"),
		net.ParseIP("176.31.205.156"),
		net.ParseIP("176.31.205.158"),
	}

	netIps := convertIPsToShortMode(ips, 40)
	fmt.Printf("%+v\n", netIps)
}

func convertIPsToShortMode(IPs []net.IP, precision uint) string {

	// Deixa apenas IPs únicos
	// fmt.Println("Entrada de IPs para conversão em modo de encurtamento", IPs)
	// cria um mapa de chave de strings e valor booleano, o
	// tamanho do mapa é de acordo o retorno do len(IP), .a quantidade de IPs
	ipsMap := make(map[string]bool, len(IPs))

	// mapa do tipo []net.IP
	// https://golang.org/pkg/builtin/#make
	IPsUniq := make([]net.IP, 0, len(IPs))
	for _, ip := range IPs {
		if _, ok := ipsMap[ip.String()]; !ok {
			ipsMap[ip.String()] = true
			IPsUniq = append(IPsUniq, ip)
		}
	}
	// Ordena o array para iniciar os processos de validações
	sort.Slice(IPsUniq, func(i, j int) bool {
		return binary.BigEndian.Uint32(IPsUniq[i].To4()) < binary.BigEndian.Uint32(IPsUniq[j].To4())
	})

	netIps, _ := getPercent(IPsUniq)

	var stringOut string
	for _, rangeIP := range netIps {
		// Se só for um ip vamos omitir o /32 para diminuir a string
		// segundo a RFC 7208 sessão 5.6 isso é permitido
		// https://tools.ietf.org/html/rfc7208#section-5.6
		stringOut += " ip4:" + strings.TrimSuffix(rangeIP.String(), "/32")
	}

	stringOut = strings.TrimSpace(stringOut)
	//

	// Step 2. Convert to binary
	// Step 3. Calculate the total number of hosts per subnet
	// Step 4. Calculate the number of subnets

	return stringOut
}


func getPercent(ips []net.IP) ([]*net.IPNet, []net.IP) {
	var ipsToFloatList []float64
	var ipsList  []net.IP
	var ipsRanges [][]net.IP
	var currentRange int = -1

	// Transforma os ips em uint
	for _, ip := range ips {
		if ip != nil {
			ipsToFloatList = append(ipsToFloatList, float64(ipToInt(ip)))
		}
	}

	// IPs que serão retornados
	var returnIPs []net.IP
	for _, ip := range ipsToFloatList {
		returnIPs = append(returnIPs, net.IP(getBinary(uint32(ip))))
	}

	// vamos iterar sobre os ips
	for idx, ipFloat := range ipsToFloatList {

		ip := net.IP(getBinary(uint32(ipFloat))).To4()
		// Step 1. Find host range
		// Vamos agrupar por ranges
		if currentRange == -1 {
			currentRange = int(ip[2])
		}

		// pega o range do ip
		if int(ip[2]) == currentRange {
			ipsList = append(ipsList, ip)
			if idx == len(ipsToFloatList)-1 {
				ipsRanges = append(ipsRanges, ipsList)
			}
			continue
		}

		ipsRanges = append(ipsRanges, ipsList)
		ipsList = nil
		ipsList = append(ipsList, ip)
		currentRange = int(ip[2])
	}

	// IPNets que serão retornados
	var returnIPNet []*net.IPNet
	for _, list := range ipsRanges {

		// vamos trabalhar em cima desse calculo, pois ele será baseado em nossa precisão
		t := int((list[len(list)-1])[3]-(list[0])[3]) + 1

		var ipnets []*net.IPNet
		ipnets = append(ipnets, GetCidr(list[0].String(), int(t)))
		returnIPNet = append(returnIPNet, ipnets...)
		continue
	}

	return returnIPNet, returnIPs
}

func getBinary(data interface{}) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, data)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func ipToInt(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func GetCidr(ipstr string, total int) *net.IPNet {
	fmt.Println(total)
	ip := net.ParseIP(ipstr).To4()
	ipMax := net.ParseIP(fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3]+byte(total)))
	bar := uint32(33 - math.Log2(float64(total)) - 1)


	q := ip.String() + fmt.Sprintf("/%d", bar)

	_, ipnet, _ := net.ParseCIDR(q)

	for cidr := bar; cidr > 0; cidr-- {
		if ipnet.IP[3] < ip[3] {
			q = ipnet.IP.String() + fmt.Sprintf("/%d", cidr)
			_, ipnet, _ = net.ParseCIDR(q)
			if ipnet.Contains(ipMax) {
				break
			}
		} else {
			break
		}
	}

	return ipnet
}